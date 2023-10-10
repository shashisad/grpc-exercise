package main

import (
	"context"
	"fmt"
	"gitlab.eng.vmware.com/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("client")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect : %v", err)
	}
	defer conn.Close()
	c := greetpb.NewGreetServiceClient(conn)
	//fmt.Println("created client: %f", c)
	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	doBiDirectionalStreaming(c)
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do Unary rpc")
	req := &greetpb.GreetingRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Shashi",
			LastName:  "Sad",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet rpc : %v", err)
	}

	log.Println("Response from Greet: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do GreetManyTimes streaming rpc")
	req := &greetpb.GreetingManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Shashi",
			LastName:  "Sad",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling streaming  GreetManyTime RPC : %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			//we have reached at the end of stream
			break
		}

		if err != nil {
			log.Fatalf("error while reading stream : %v", err)
		}

		log.Println("Response from GreetManyTimes : %v", msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do LongGreet streaming rpc")
	requests := []*greetpb.LongGreetingRequest{
		&greetpb.LongGreetingRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Shashi",
			},
		},
		&greetpb.LongGreetingRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Priyanshi",
			},
		},
		&greetpb.LongGreetingRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Nidhi",
			},
		},
		&greetpb.LongGreetingRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Shreyansh",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalln("error while calling LongGreet : %v", err)
	}
	//we iterate over slice and send each req individually
	for _, req := range requests {
		fmt.Println("Sending request : %v", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalln("error while receiving response")
	}
	fmt.Println("LongGreet Response : %v", res)
}

func doBiDirectionalStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do GreetEveryone  BiDi-streaming rpc")

	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Shashi",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Priyanshi",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Nidhi",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Shreyansh",
			},
		},
	}

	// we create a stream by invoking the client
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalln("Error while creating stream : %v", err)
		return
	}

	waitc := make(chan struct{})

	// we send a bunch of messages to the client(go routine)
	go func() {
		// function to  send a bunch of messages
		for _, req := range requests {
			fmt.Println("Sending message : %v", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	// we receive a bunch of messages from the client(go routine)
	go func() {
		// function to receive a bunch of messages
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				break
			}

			if err != nil {
				log.Fatalln("Error while receiving: %v", err)
				close(waitc)
				break
			}
			fmt.Println("Received : %v", res)
		}

	}()

	// block until everything is fine
	<-waitc
}
