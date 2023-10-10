package main

import (
	"context"
	"fmt"
	"gitlab.eng.vmware.com/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("calculator client")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect : %v", err)
	}
	defer conn.Close()
	c := calculatorpb.NewCalculatorServiceClient(conn)
	//fmt.Println("created client: %f", c)
	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	doBiDiStreaming(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do Sum Unary rpc......")
	req := &calculatorpb.SumRequest{
		FirstNumber:  5,
		SecondNumber: 56,
	}
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Sum rpc : %v", err)
	}

	log.Println("Response from  Sum: %v", res.SumResult)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do prime server streaming rpc......")
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 12346578790,
	}
	stream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling PrimeNumberDecompostion (server streaming) rpc : %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			//we have reached at the end of stream
			break
		}

		if err != nil {
			log.Fatalf("error while reading stream : %v", err)
		}

		log.Println(res.GetPrimeFactor())
	}
}
func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do Average streaming rpc")
	//requests := []*calculatorpb.ComputeAverageRequest{
	//	&calculatorpb.ComputeAverageRequest{
	//		Number: int32(2),
	//	},
	//	&calculatorpb.ComputeAverageRequest{
	//		Number: int32(2),
	//	},
	//	&calculatorpb.ComputeAverageRequest{
	//		Number: int32(2),
	//	}, &calculatorpb.ComputeAverageRequest{
	//		Number: int32(2),
	//	}, &calculatorpb.ComputeAverageRequest{
	//		Number: int32(2),
	//	},
	//}

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalln("error while calling ComputeAverage : %v", err)
	}
	numbers := []int32{3, 5, 9, 54, 23}
	//we iterate over slice and send each req individually
	for _, number := range numbers {
		fmt.Println("Sending request : %v", number)
		stream.Send(&calculatorpb.ComputeAverageRequest{
			Number: number,
		})
		time.Sleep(100 * time.Millisecond)
	}

	//for _, req := range requests {
	//	fmt.Println("Sending request : %v", req)
	//	stream.Send(req)
	//	time.Sleep(100 * time.Millisecond)
	//}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalln("error while receiving response : %v", err)
	}
	fmt.Println("ComputeAverage Response : %v", res.GetAverage())
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do Bi directional streaming rpc")

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalln("Error while opening stream and calling FindMaximum : %v", err)
	}

	waitc := make(chan struct{})

	//send go routine
	go func() {
		numbers := []int32{4, 7, 19, 4, 6, 32}
		for _, number := range numbers {
			stream.Send(&calculatorpb.FindMaximumRequest{
				Number: number,
			})
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	//receive ro routine
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalln("Error while reading server stream : %v", err)
				break
			}
			maximum := res.GetMaximum()
			fmt.Println("Received a new maximum of  ... %v", maximum)
		}
		close(waitc)
	}()

	<-waitc
}
