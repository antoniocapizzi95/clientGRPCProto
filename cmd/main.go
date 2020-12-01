package main

import (
	"clientGRPCProto"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"log"
)

func getPerson(ctx context.Context, client clientGRPCProto.RPCServiceClient, number string, phoneType clientGRPCProto.PhoneType) *clientGRPCProto.Person {
	phone := &clientGRPCProto.PhoneNumber{Number: &number, Type: &phoneType}
	person, err := client.GetPersonByPhoneNumber(ctx, phone)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	fmt.Println("getPerson, name: " + *person.Name)
	return person
}
func editPeople(ctx context.Context, client clientGRPCProto.RPCServiceClient, people []clientGRPCProto.Person) {
	stream, err := client.EditPeople(ctx)
	if err != nil {
		log.Fatalf(err.Error())
	}
	for _, person := range people {
		if err := stream.Send(&person); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream, person, err)
		}
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	if *reply.Result {
		fmt.Println("People edited successfully")
	} else {
		fmt.Println("Error editing people")
	}
}

func listPeople(ctx context.Context, client clientGRPCProto.RPCServiceClient, number *clientGRPCProto.PhoneNumber) {
	stream, err := client.ListPeopleByPhoneType(ctx, number)
	if err != nil {
		fmt.Println("error with stream, " + err.Error())
		return
	}
	for {
		person, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListPeople(_) = _, %v", client, err)
		}
		fmt.Println(*person.Name)
	}
}
func getPeopleById(ctx context.Context, client clientGRPCProto.RPCServiceClient, ids []clientGRPCProto.RequestId) {
	stream, err := client.GetPeopleById(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a request : %v", err)
			}
			log.Printf("Got person %s", *in.Name)
		}
	}()
	for _, id := range ids {
		if err := stream.Send(&id); err != nil {
			log.Fatalf("Failed to send a request: %v", err)
		}
	}
	stream.CloseSend()
	<-waitc
}
func main() {
	var conn *grpc.ClientConn
	ctx := context.Background()
	conn, err := grpc.Dial("10.8.0.1:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()
	client := clientGRPCProto.NewRPCServiceClient(conn)
	number := "4365365432"
	phoneType := clientGRPCProto.PhoneType_HOME

	fmt.Println("--------- GET A PERSON WITH NUMBER: " + number + " ----------")
	person := getPerson(ctx, client, number, phoneType)

	fmt.Println("-------- Editing John Doe in Giovanni Doe and Mario Rossi in Mario Bianchi --------")
	newName := "Giovanni Doe"
	person.Name = &newName
	number = "452376467"
	person2 := getPerson(ctx, client, number, phoneType)
	newName2 := "Mario Bianchi"
	person2.Name = &newName2
	editPeople(ctx, client, []clientGRPCProto.Person{*person, *person2})
	getPerson(ctx, client, number, phoneType)

	fmt.Println("------ Listing People with at least a number of type HOME --------")
	listPeople(ctx, client, &clientGRPCProto.PhoneNumber{Number: &number, Type: &phoneType})

	fmt.Println("------- Get People with ID 1 and 2 ---------")
	var id1 int32 = 1
	var id2 int32 = 2
	reqId1 := clientGRPCProto.RequestId{Id: &id1}
	reqId2 := clientGRPCProto.RequestId{Id: &id2}
	ids := []clientGRPCProto.RequestId{reqId1, reqId2}
	ids[0].Id = &id1
	ids[1].Id = &id2
	getPeopleById(ctx, client, ids)

}
