package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func main() {
	// Токен из переменной окружения
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Установите переменную окружения BOT_TOKEN")
	}

	// Создаём бота
	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	me, _ := bot.GetMe(context.Background())
	fmt.Printf("Бот запущен: @%s (%s)\n", me.Username, me.FirstName)

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Long polling
	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		log.Fatalf("Ошибка long polling: %v", err)
	}

	// Обработчик обновлений
	handler, err := th.NewBotHandler(bot, updates)
	if err != nil {
		log.Fatalf("Ошибка создания обработчика: %v", err)
	}
	defer handler.Stop()

	// Обрабатываем все сообщения
	handler.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		chatID := tu.ID(message.Chat.ID)

		// Если есть текст - отправляем текст обратно
		if message.Text != "" {
			params := tu.Message(chatID, message.Text)

			_, err := bot.SendMessage(ctx, params)
			return err
		}

		return err
	})

	// Запуск обработки
	go handler.Start()

	fmt.Println("Бот работает... Нажмите Ctrl+C для остановки")

	// Ожидание сигнала завершения
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	cancel() // Останавливаем polling
	fmt.Println("\nБот остановлен")
}
