package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func TestUI(t *testing.T) {
	// Поднимаем локальный сервер для раздачи нашего index.html
	fs := http.FileServer(http.Dir("."))
	server := httptest.NewServer(fs)
	defer server.Close()

	// Настраиваем подключение к ChromeDriver
	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService("chromedriver", 4444, opts...)
	if err != nil {
		t.Fatalf("Ошибка запуска ChromeDriver: %v", err)
	}
	defer service.Stop()

	// Настраиваем Chrome для работы в фоновом режиме (Headless mode) - это нужно для CI/CD
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{"--headless", "--no-sandbox", "--disable-dev-shm-usage"},
	}
	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 4444))
	if err != nil {
		t.Fatalf("Ошибка подключения к WebDriver: %v", err)
	}
	defer wd.Quit()

	// Открываем нашу страницу
	if err := wd.Get(server.URL + "/index.html"); err != nil {
		t.Fatalf("Не удалось загрузить страницу: %v", err)
	}

	// ТЕСТ 1: Проверка заголовка окна (Title)
	title, _ := wd.Title()
	if title != "Test Form" {
		t.Errorf("Ожидался title 'Test Form', получено '%s'", title)
	}

	// ТЕСТ 2: Проверка заголовка H1
	headerElem, err := wd.FindElement(selenium.ByID, "header")
	if err != nil {
		t.Fatalf("Не удалось найти заголовок H1: %v", err)
	}
	headerText, _ := headerElem.Text()
	if headerText != "Регистрация" {
		t.Errorf("Ожидался заголовок 'Регистрация', получено '%s'", headerText)
	}

	// ТЕСТ 3: Проверка ввода в форму
	inputElem, _ := wd.FindElement(selenium.ByID, "username")
	inputElem.SendKeys("Иван Иванов")

	btnElem, _ := wd.FindElement(selenium.ByID, "submitBtn")
	btnElem.Click()

	time.Sleep(500 * time.Millisecond) // Немного ждем отработки JavaScript

	// ТЕСТ 4: Проверка текста после отправки формы
	resultElem, _ := wd.FindElement(selenium.ByID, "result")
	resultText, _ := resultElem.Text()
	if resultText != "Успешно!" {
		t.Errorf("Ожидался текст 'Успешно!', получено '%s'", resultText)
	}
}
