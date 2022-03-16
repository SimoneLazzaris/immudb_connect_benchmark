package main

import (
	"context"
	"flag"
	"github.com/codenotary/immudb/pkg/api/schema"
	immuclient "github.com/codenotary/immudb/pkg/client"
	"google.golang.org/grpc/metadata"
	"log"
	"math/rand"
	"time"
)

var config struct {
	IpAddr    string
	Port      int
	Username  string
	Password  string
	DBName    string
	Workers   int
	Batchsize int
	Batchwait int
	Loopsize  int
	Loopwait  int
}

func init() {
	log.SetFlags(log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	flag.StringVar(&config.IpAddr, "addr", "", "IP address of immudb server")
	flag.IntVar(&config.Port, "port", 3322, "Port number of immudb server")
	flag.StringVar(&config.Username, "user", "immudb", "Username for authenticating to immudb")
	flag.StringVar(&config.Password, "pass", "immudb", "Password for authenticating to immudb")
	flag.StringVar(&config.DBName, "db", "", "Name of the database to use")
	flag.IntVar(&config.Workers, "workers", 1, "Concurrent workers")
	flag.IntVar(&config.Batchsize, "batchsize", 1, "Iteration per workers")
	flag.IntVar(&config.Batchwait, "batchwait", 100, "Average sleep time between batches")
	flag.IntVar(&config.Loopsize, "loopsize", 1, "Tight loop iteration per workers")
	flag.IntVar(&config.Loopwait, "loopwait", 10, "Average sleep time inside loop")
	flag.Parse()
}

func work(n, i int) {
	log.Printf("Client %d:%d starting", n, i)
	opts := immuclient.DefaultOptions().WithAddress(config.IpAddr).WithPort(config.Port)

	client, err := immuclient.NewImmuClient(opts)
	if err != nil {
		log.Fatalln("Failed to connect. Reason:", err)
	}
	ctx := context.Background()

	login, err := client.Login(ctx, []byte(config.Username), []byte(config.Password))
	if err != nil {
		log.Fatalln("Failed to login. Reason:", err.Error())
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", login.GetToken()))
	defer func() {
		client.Disconnect()
	}()
	if config.DBName != "" {
		udr, err := client.UseDatabase(ctx, &schema.Database{DatabaseName: config.DBName})
		if err != nil {
			log.Fatalln("Failed to use the database. Reason:", err)
		}
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", udr.GetToken()))
	}
	for j := 0; j < config.Loopsize; j++ {
		client.Health(ctx)
		client.CurrentState(ctx)
		if config.Loopwait>0 {
			delay := rand.Intn(config.Loopwait) + config.Loopwait/2
			time.Sleep(time.Millisecond * time.Duration(delay))
		}
	}
	log.Printf("Client %d end", n)
}

func main() {
	end := make(chan bool)
	for i := 0; i < config.Workers; i++ {
		go func(c int) {
			for j := 0; j < config.Batchsize; j++ {
				work(c+1, j)
				if config.Batchwait > 0 {
					delay := rand.Intn(config.Batchwait) + config.Batchwait/2
					time.Sleep(time.Millisecond * time.Duration(delay))
				}
			}
			end <- true
		}(i)
	}
	for i := 0; i < config.Workers; i++ {
		<-end
	}
}
