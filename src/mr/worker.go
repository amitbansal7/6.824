package mr

import "fmt"
import "log"
import "net/rpc"
import "os"
import "hash/fnv"

import "encoding/json"

// import "sync"
// import "math/rand"
import "io/ioutil"
import "strconv"
import "time"
import "sort"

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// var seededRand *rand.Rand = rand.New(
// rand.NewSource(int(time.Now().UnixNano() / 1e6))

type meta struct {
	work    *Work
	wChan   chan *Work
	NReduce int
	MapDone bool
	AllDone bool
	mapf    func(string, string) []KeyValue
	reducef func(string, []string) string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

func (m *meta) Sync() {
	for {

		response := SyncResponse{}

		if time.Now().After(m.work.Timeout) {
			// fmt.Println("worker Timeout marking worker as idle")
			m.work.Status = "idle"
		}
		if !call("Master.Sync", &m.work, &response) {
			//s.Init(s.mapf,s.reducef)
			fmt.Println("os.Exit(0) from worker")
			os.Exit(0)
		}

		if response.NewWork != nil {
			m.work = response.NewWork
			m.wChan <- response.NewWork
		}

		m.AllDone = response.AllDone
		m.MapDone = response.MapDone
		m.NReduce = response.NReduce

		if m.AllDone {
			// fmt.Println("All Done!!!")
			close(m.wChan)
			return
		}

		time.Sleep(time.Millisecond * 500)
	}
}

func (m *meta) Map() {
	// fmt.Println("Map -> ", m.work.Task.File)
	// defer fmt.Println("Map done-> ", m.work.Task.File)

	task := m.work.Task
	file, err := os.Open(task.File)
	if err != nil {
		fmt.Println("Cannot open file", file)
	}
	content, err := ioutil.ReadAll(file)

	if err != nil {
		fmt.Println("Reading error", file)
	}

	file.Close()

	kvs := m.mapf(task.File, string(content))

	var fileToKvs map[string][]KeyValue

	fileToKvs = map[string][]KeyValue{}

	for _, kv := range kvs {
		outf := "mr-reduce-in-" + strconv.Itoa(ihash(kv.Key)%m.NReduce+1)

		if _, ok := fileToKvs[outf]; !ok {
			fileToKvs[outf] = []KeyValue{}
		}

		fileToKvs[outf] = append(fileToKvs[outf], kv)
	}

	for out, kvs := range fileToKvs {
		tempFile := strconv.FormatInt(time.Now().UnixNano(), 10)
		if _, e := os.Create(tempFile); e != nil {
			log.Printf("[Master] Create [%s] File Error", tempFile)
		}
		f, err := os.OpenFile(tempFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

		if err != nil {
			fmt.Println("Cannot open file ", tempFile)
		}
		encoder := json.NewEncoder(f)
		for _, kv := range kvs {
			enc := encoder.Encode(&kv)
			if enc != nil {
				fmt.Println("enc error ", enc)
			}
		}

		task.TempToResFiles[tempFile] = out

		f.Close()
	}
}

func (m *meta) Reduce() {
	// fmt.Println("Reduce -> ", m.work.Task.File)
	// defer fmt.Println("Reduce done -> ", m.work.Task.File)

	task := m.work.Task
	file, err := os.Open(task.File)
	if err != nil {
		fmt.Println("Cannot open file", file)
	}
	// content, err := ioutil.ReadAll(file)

	decoder := json.NewDecoder(file)
	intermediate := []KeyValue{}

	for {
		var kv KeyValue
		if err := decoder.Decode(&kv); err != nil {
			break
		}
		intermediate = append(intermediate, kv)
	}

	sort.Sort(ByKey(intermediate))

	var fileToKvs map[string][]KeyValue

	fileToKvs = map[string][]KeyValue{}

	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := m.reducef(intermediate[i].Key, values)

		outf := "mr-out-" + strconv.Itoa(ihash(intermediate[i].Key)%m.NReduce+1)

		if _, ok := fileToKvs[outf]; !ok {
			fileToKvs[outf] = []KeyValue{}
		}

		fileToKvs[outf] = append(fileToKvs[outf], KeyValue{Key: intermediate[i].Key, Value: output})

		i = j
	}

	for out, kvs := range fileToKvs {
		tempFile := strconv.FormatInt(time.Now().UnixNano(), 10)
		if _, e := os.Create(tempFile); e != nil {
			log.Printf("[Master] Create [%s] File Error", tempFile)
		}
		f, err := os.OpenFile(tempFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			fmt.Println("Cannot open file ", tempFile)
		}

		encoder := json.NewEncoder(f)
		for _, kv := range kvs {
			enc := encoder.Encode(&kv)
			if enc != nil {
				fmt.Println("enc error ", enc)
			}
		}

		// for _, kv := range kvs {
		// 	fmt.Fprintf(f, "%v %v\n", kv.Key, kv.Value)
		// }
		task.TempToResFiles[tempFile] = out
		f.Close()
	}

}

func (m *meta) DoWork() {
	for work := range m.wChan {
		if t := work.Task; t != nil {
			if t.Action == "map" {
				m.Map()
				m.work.Status = "done"
			} else if t.Action == "reduce" {
				if m.MapDone == false {
					// fmt.Println("Reduce task received waiting for map to complete...")
					time.Sleep(time.Millisecond * 300)
					m.wChan <- work
				} else {
					m.Reduce()
					m.work.Status = "done"
				}
			}

		}
	}
}

func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	m := meta{}

	m.work = &Work{
		Status:  "idle",
		Timeout: time.Now(),
	}

	m.wChan = make(chan *Work, 2)

	m.mapf = mapf
	m.reducef = reducef

	go m.Sync()

	m.DoWork()

	// fmt.Println("Worker shutting down.....")
}

func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
