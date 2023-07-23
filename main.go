package main

import (
	"fmt"
	//"sync"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Ttype represents a meaninglessness of our life
type Ttype struct {
	id         int
	createdTime         string // время создания
	finishedTime         string // время выполнения
	taskRESULT []byte
}

func taskCreturer(cancel <-chan struct{}) <-chan Ttype {
	superChan := make(chan Ttype)
	go func() {
		defer close(superChan)
		for {
			createdTime := time.Now().Format(time.RFC3339)
			if time.Now().UnixMicro()%2 > 0 { // вот такое условие появления ошибочных тасков
				createdTime = "Some error occured"
			}
			//fmt.Println(createdTime)
			select {
			case superChan <- Ttype{createdTime: createdTime, id: int(time.Now().Unix())}: // передаем таск на выполнение"
			case <-cancel:
				return
			}
		}
	}()
	return superChan
}

func task_worker(t Ttype) Ttype {
	tt, err := time.Parse(time.RFC3339, t.createdTime)
	if err != nil {
		t.taskRESULT = []byte("something went wrong")
	} else {
		if tt.After(time.Now().Add(-20 * time.Second)) {
			t.taskRESULT = []byte("task has been successed")
		} else {
			t.taskRESULT = []byte("something went wrong")
		}
	}

	t.finishedTime = time.Now().Format(time.RFC3339Nano)

	time.Sleep(time.Millisecond * 150)

	return t
}

func tasksorter(in <-chan Ttype, cancel <-chan struct{}) (<-chan Ttype, <-chan error) {
	doneTasks := make(chan Ttype)
	undoneTasks := make(chan error)
	go func() {
		defer close(doneTasks)
		defer close(undoneTasks)
		for val := range in {
			val := val
	//		go func() {
			res := task_worker(val)
			if string(res.taskRESULT[14:]) == "successed" {
				select {
				case doneTasks <- res:
				case <-cancel:
					return
				}

			} else {
				select {
				case undoneTasks <- fmt.Errorf("task id %d time %s, error %s", res.id, res.finishedTime, res.taskRESULT):
				case <-cancel:
					return
				}

			}
		//	}()
		}
	}()
	return doneTasks, undoneTasks
}
func main() {
	cancel := make(chan struct{})
	tc := taskCreturer(cancel)
	done, undone := tasksorter(tc, cancel)

	
	result := map[int]Ttype{}
	err := []error{}
	go func() {
		for r := range done {
			r := r
			result[r.id] = r

		}
	}()
	go func() {
		for r := range undone {
			r := r
			err = append(err, r)
		}
	}()
	time.Sleep(time.Second * 4)
	close(cancel)
	println("Errors:")
	for _, i := range err {
		fmt.Println(i)
	}

	println("Done tasks:")
	for r := range result {
		fmt.Println(r)
	}

}
