package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type ClientInfo struct {
	LastPing    time.Time //время последнего пинга клиета
	Status      string    //статус клиента
	ChangeCount int       //счетчки изменения статуса
}

var clients = make(map[string]*ClientInfo) //мапа для хранения данных клиента
var mutex = sync.Mutex{}                   //мьютекс для доступа к мапе
var checkPeriod = 1 * time.Minute          // период проверки статуса клиента

// обновление файла со статусом
func updateStatusFiles() {
	statusFile, _ := os.Create("status.txt") //создание файла status.txt
	defer statusFile.Close()                 //закрытие файла в конце функ

	//запись статуса
	for id, client := range clients {
		fmt.Fprintf(statusFile, "%s, %s\n", id, client.Status)
	}
}

// функ для обработки запросов на пинг
func pingHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Path[len("/ping/"):] //экстрек клиентИд из юрла
	mutex.Lock()                           //блок
	defer mutex.Unlock()                   //по завершеник функ осбождение мьютекса

	//клиент уже есть в мапе?
	if client, exists := clients[clientID]; exists {
		if client.Status == "offline" { // если он етсь ставим его онфлайн
			client.Status = "online"
			client.ChangeCount++ //увел счет
		}
		client.LastPing = time.Now() //время последнего пинга
	} else {
		clients[clientID] = &ClientInfo{ //создаем клиента
			LastPing:    time.Now(),
			Status:      "online",
			ChangeCount: 0,
		}
	}

	fmt.Fprintf(w, "Сигнал, полученный от %s", clientID)
	updateStatusFiles() //обновили файл статуса
}

// мониторинг статусов
func monitorClients() {
	for {
		time.Sleep(checkPeriod) // пауза между проверками
		mutex.Lock()            //блок
		for id, client := range clients {
			//с последнего пинга прошло больше времени?
			if time.Since(client.LastPing) > checkPeriod {
				if client.Status == "online" {
					client.Status = "offline" //ставим офлайн
					client.ChangeCount++
					//если статус менялся 5 раз записываем ошибку
					if client.ChangeCount >= 5 {
						recordError(id)
						client.ChangeCount = 0
					}
				}
			}
		}
		mutex.Unlock()
		updateStatusFiles()
	}
}

// запись ошибки
func recordError(clientID string) {
	file, err := os.OpenFile("status-error.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return //выход из функ при ошибки открытия файла
	}
	defer file.Close() //закрытие в конце

	now := time.Now().Format("2006-01-02, 15:04:05")                                 //текущее время
	_, err = file.WriteString(fmt.Sprintf("%s, Network error, %s\n", clientID, now)) //запись ошибки
	if err != nil {
		fmt.Println("ошибка записи в журнал ошибок: ", err)
	}
}

func main() {
	http.HandleFunc("/ping/", pingHandler)
	go monitorClients()

	fmt.Println("Сервер запущен на порту :8088")
	http.ListenAndServe(":8088", nil)
}
