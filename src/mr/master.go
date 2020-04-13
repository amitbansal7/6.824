package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "time"
import "net/http"
import "encoding/json"

// import "errors"
import "sync"
import "strconv"
import "fmt"

// import "io/ioutil"

type Master struct {
	activeWorks       map[string]*Work
	mapTasks          []Task
	reduceTasks       []Task
	mapActiveWorks    int
	reduceActiveWorks int
	nReduce           int
	mu                sync.Mutex
	filesMu           sync.Mutex
	started           bool
	allDone           bool
}

func (m *Master) Init(files []string, nReduce int) {
	m.activeWorks = map[string]*Work{}
	m.mapActiveWorks = 0
	m.reduceActiveWorks = 0
	m.mapTasks = []Task{}
	m.reduceTasks = []Task{}
	m.nReduce = nReduce
	m.started = true
}

var (
	startTime time.Time
)

func (m *Master) CreateTasks(files []string) {
	reduceTaskFiles := []string{}

	for i := 0; i < m.nReduce; i++ {
		out := "mr-out-" + strconv.Itoa(i+1)
		reduceFile := "mr-reduce-in-" + strconv.Itoa(i+1)
		reduceTaskFiles = append(reduceTaskFiles, reduceFile)

		if _, e := os.Create(out); e != nil {
			log.Printf("[Master] Create [%s] File Error", out)
		}

		if _, e := os.Create(reduceFile); e != nil {
			log.Printf("[Master] Create [%s] File Error", reduceFile)
		}
	}

	m.mu.Lock()
	m.AddTasks("map", files...)
	m.AddTasks("reduce", reduceTaskFiles...)
	m.mu.Unlock()
}

//action = map or reduce
func (m *Master) AddTasks(action string, files ...string) {
	var tasks []Task

	for _, file := range files {
		tasks = append(tasks, Task{
			Action:         action,
			File:           file,
			TempToResFiles: map[string]string{},
		})
	}

	if action == "map" {
		m.mapTasks = append(tasks, m.mapTasks...)
	} else {
		m.reduceTasks = append(m.reduceTasks, tasks...)
	}
}

func (m *Master) RemoveActiveWork(work *Work) {
	// m.mu.Lock()
	// defer m.mu.Unlock()

	if _, ok := m.activeWorks[work.Id]; ok {
		delete(m.activeWorks, work.Id)
		if work.Task.Action == "map" {
			m.mapActiveWorks -= 1
		} else if work.Task.Action == "reduce" {
			m.reduceActiveWorks -= 1
		}
	}
}

func (m *Master) Checker() {
	for {

		// fmt.Println("Current Status: ",
		// "activeWorks : ", len(m.activeWorks),
		// "| mapTasks : ", len(m.mapTasks),
		// "| reduceTasks : ", len(m.reduceTasks),
		// "| mapActiveWorks : ", m.mapActiveWorks,
		// "| reduceActiveWorks : ", m.reduceActiveWorks,
		// )

		if m.allDone {
			return
		}
		for _, w := range m.activeWorks {
			if time.Now().After(w.Timeout) {
				// fmt.Println("Work timeout... reassigning task")
				m.mu.Lock()
				m.RemoveActiveWork(w)
				m.AddTasks(w.Task.Action, w.Task.File)
				m.mu.Unlock()

			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (m *Master) CreateNewWork() (newWork *Work) {
	// m.mu.Lock()
	// defer m.mu.Unlock()

	var task *Task

	if len(m.mapTasks) > 0 {
		task = &m.mapTasks[0]
		m.mapTasks = m.mapTasks[1:]
		m.mapActiveWorks += 1
	} else if len(m.reduceTasks) > 0 {
		task = &m.reduceTasks[0]
		m.reduceTasks = m.reduceTasks[1:]
		m.reduceActiveWorks += 1
	}

	var work Work

	if task != nil {
		work = Work{
			Id:      fmt.Sprintf("%s - %s", task.File, task.Action),
			Status:  "working",
			Timeout: time.Now().Add(time.Second * 10),
			Task:    task,
		}
		newWork = &work
		m.activeWorks[work.Id] = &work
	}

	return &work
}

func (m *Master) UpdateFiles(task *Task) {
	m.filesMu.Lock()
	defer m.filesMu.Unlock()
	if task.Action == "map" {
		for t, o := range task.TempToResFiles {
			temp, err := os.Open(t)
			if err != nil {
				fmt.Println("Cannot open file", t)
			}
			out, err := os.OpenFile(o, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			if err != nil {
				fmt.Println("Cannot open file", o)
			}

			tempDecoder := json.NewDecoder(temp)
			outEncoder := json.NewEncoder(out)

			for {
				var kv KeyValue
				if err := tempDecoder.Decode(&kv); err != nil {
					break
				}
				enc := outEncoder.Encode(&kv)
				if enc != nil {
					fmt.Println("enc error ", enc)
				}
			}
			temp.Close()
			out.Close()
			err = os.Remove(t)
			if err != nil {
				fmt.Println(err)
			}
		}
	} else if task.Action == "reduce" {
		for t, o := range task.TempToResFiles {
			temp, err := os.Open(t)
			if err != nil {
				fmt.Println("Cannot open file", t)
			}
			out, err := os.OpenFile(o, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
			if err != nil {
				fmt.Println("Cannot open file", o)
			}
			tempDecoder := json.NewDecoder(temp)

			for {
				var kv KeyValue
				if err := tempDecoder.Decode(&kv); err != nil {
					break
				}
				fmt.Fprintf(out, "%v %v\n", kv.Key, kv.Value)
			}

			temp.Close()
			out.Close()

			err = os.Remove(t)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (m *Master) Sync(work *Work, response *SyncResponse) error {

	// fmt.Println("Worker asking for Sync worker status: ", work.Status)

	switch work.Status {

	case "idle":
		m.mu.Lock()
		m.RemoveActiveWork(work)
		response.NewWork = m.CreateNewWork()
		m.mu.Unlock()

	case "done":
		//If work is active and not removed by Checker due to timeout
		if _, ok := m.activeWorks[work.Id]; ok {
			// fmt.Println("Work Done!!", work.Task)
			m.mu.Lock()
			m.UpdateFiles(work.Task)
			m.RemoveActiveWork(work)
			response.NewWork = m.CreateNewWork()
			m.mu.Unlock()
		} else {
			// fmt.Println("******Dead worker is alive again", work.Task)
		}

	case "working":
		//do nothing..
	}

	response.MapDone = len(m.mapTasks) == 0 && m.mapActiveWorks == 0
	response.AllDone = response.MapDone && len(m.reduceTasks) == 0 && m.reduceActiveWorks == 0
	m.allDone = response.AllDone

	response.NReduce = m.nReduce
	return nil
}

func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (m *Master) Done() bool {
	ret := false
	if m.mapActiveWorks == 0 && m.reduceActiveWorks == 0 && len(m.mapTasks) == 0 && len(m.reduceTasks) == 0 && m.started {
		// fmt.Println("Master Done in", time.Now().Sub(startTime).Seconds(), "Seconds")
		ret = true
	}

	return ret
}

func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	startTime = time.Now()

	m.Init(files, nReduce)
	m.CreateTasks(files)
	// fmt.Println("Init done")
	go m.Checker()

	m.server()
	return &m
}
