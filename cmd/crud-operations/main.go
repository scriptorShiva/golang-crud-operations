package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scriptorshiva/golang-crud-operations/internal/config"
	"github.com/scriptorshiva/golang-crud-operations/internal/http/handlers/student"
	"github.com/scriptorshiva/golang-crud-operations/internal/storage/sqlite"
)

func main() {
	/*
		1. Load config
		2. DB setup
		3. Setup Router
		4. Setup Server
	*/

	// load config
	cfg := config.MustLoad()

	// DB setup
	storage,err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storage initialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))

	
	// router setup
	router := http.NewServeMux()

	// route handler
	// we are giving student reference to handler , and we can use DI also by passing references to Student.new()
	// before passing storage we need to implements the interface. But, in go we don't have "implemnents". Go handles that internally
	router.HandleFunc("POST /api/student", student.New(storage))
	router.HandleFunc("GET /api/student/{id}", student.FetchById(storage))
	router.HandleFunc("GET /api/students", student.FetchAll(storage))

	// setup server
	server := http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler: router,
	}

	fmt.Println("Server is up and running")

	// ---- Normal server exit ----
	// err := server.ListenAndServe()
	// if err != nil {
	// 	// log.Fatal("Failed to start server")
	// 	panic(err)
	// }

	// ---- We should always Gracefully shutdown server : Stop accepting new requests, finish in-flight requests, then exit cleanly. ----
	// without gracefull shutdown - active HTTP requests get killed - DB writes can be half-done - clients see connection resets
	/*
		Start server
			↓
		Wait for OS signal (Ctrl+C / SIGTERM)
			↓
		Tell server: "stop accepting new requests"
			↓
		Wait up to 5s for ongoing requests to finish
			↓
		Exit process cleanly

	*/
	// go routine -- concurrently running
	// problem : main routine exits before go routine exits. So, we need to wait for go routine to exit
	// We can use channels to wait for go routine to exit (synchronization). 

	// Create a signal channel
	// why?? - OS signals are asynchronous - Channels are Go’s way to wait & synchronize
	// why buffer size 1 - Prevents missing a signal if it arrives early - Best practice for signal handling
	done := make(chan os.Signal, 1)

	// “When the OS sends any of these signals, push them into done channel.”
	// os interrupt - Ctrl+C , SIGTERM - docker stop , SIGINT - interrupt from terminal , SIGKILL - cannot be caught
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	go func(){
		// why go routine??
		// - ListenAndServe() blocks forever , - If you call it directly → main thread is stuck , - You wouldn’t be able to listen for signals
		// “Run the server in background, I’ll wait for signals.”
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("Failed to start server")
		}
	}()

	// “Do NOTHING until a signal arrives.” - Main goroutine sleeps here.
	// entire flow :: OS sends: SIGINT or SIGTERM , Go runtime:signal.Notify → done channel, now- done , unblocks
	<-done

	// This is mostly for: logs, observability
	slog.Info("Shutting down the server")

	// Create a shutdown context with timeout
	/*
		- Creates a deadline : - server has 5 seconds , - after that → force shutdown
		- This protects you from : - stuck requests , - hung connections , - infinite waits
	*/
	ctx,cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	// Gracefully shut down the server
	// What Shutdown() does internally
	// - Stops accepting new connections, - Keeps existing connections alive, - Waits until: all requests finish OR context times out , - Closes everything , - Returns, - If requests finish within 5s -> clean exit, - If not -> force exit
	server.Shutdown(ctx)
}

/*
	Why this called gracefull shutdown?
	Without
	- Connections dropped
	- Clients see errors
	- Data corruption risk
	- Bad for K8s

	With
	- Connections completed
	- Clients succeed
	- Safe
	- K8s-friendly

*/