
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	pb "Tarea/proto"

	"github.com/streadway/amqp"

	"google.golang.org/grpc"
)

type trevorServer struct {
	pb.UnimplementedMissionServiceServer
	currentTurns   int32
	totalTurns     int32
	isWorking      bool
	missionFailed  bool
	missionSuccess bool
	currentStars   int32
	abilityActive  bool
	baseLoot       int32
	finalLoot      int32
}

func (s *trevorServer) StartDistraction(ctx context.Context, req *pb.DistractionRequest) (*pb.DistractionResponse, error) {
	s.totalTurns = req.RequiredTurns
	s.currentTurns = 0
	s.isWorking = true
	s.missionFailed = false
	s.missionSuccess = false

	log.Printf("Trevor iniciando distracción. Turnos requeridos: %d", s.totalTurns)
	go s.workOnDistraction()

	return &pb.DistractionResponse{
		Success: true,
		Message: "Trevor comenzó la distracción",
	}, nil
}

func (s *trevorServer) StartGolpe(ctx context.Context, req *pb.GolpeRequest) (*pb.GolpeResponse, error) {
	s.totalTurns = req.RequiredTurns
	s.currentTurns = 0
	s.isWorking = true
	s.missionFailed = false
	s.missionSuccess = false
	s.baseLoot = req.BaseLoot  
	s.currentStars = 0
	s.abilityActive = false
	s.finalLoot = req.BaseLoot  

	log.Printf("Trevor iniciando golpe. Turnos requeridos: %d, Botín base: $%d", 
		s.totalTurns, s.baseLoot)

	go s.consumeStars()
	go s.workOnGolpe()

	return &pb.GolpeResponse{
		Success: true,
		Message: "Trevor comenzó el golpe",
	}, nil
}

func (s *trevorServer) consumeStars() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Printf("Error conectando a RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Error abriendo canal: %v", err)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"stars_Trevor",
		false, false, false, false, nil,
	)
	if err != nil {
		log.Printf("Error declarando cola: %v", err)
		return
	}

	msgs, err := ch.Consume(
		q.Name, "", true, false, false, false, nil,
	)
	if err != nil {
		log.Printf("Error consumiendo: %v", err)
		return
	}

	for msg := range msgs {
		stars, _ := strconv.Atoi(string(msg.Body))
		s.currentStars = int32(stars)
		log.Printf("Trevor - Estrellas actualizadas: %d", s.currentStars)

		// Habilidad especial de Trevor
		if s.currentStars >= 5 && !s.abilityActive {
			log.Println(" ¡Furia de Trevor activada! Límite aumentado a 7 estrellas")
			s.abilityActive = true
		}

		// Verificar fracaso - el límite depende de si la habilidad está activa
		maxStars := int32(5)
		if s.abilityActive {
			maxStars = 7
		}

		if s.currentStars >= maxStars {
			log.Printf(" Demasiadas estrellas (%d)! Misión fracasada", s.currentStars)
			s.missionFailed = true
			break
		}
	}
}

func (s *trevorServer) workOnDistraction() {
	for s.currentTurns < s.totalTurns && !s.missionFailed {
		time.Sleep(10 * time.Millisecond)
		s.currentTurns++

		// probabilidad de que Trevor se emborrache a la mitad
		if s.currentTurns == s.totalTurns/2 && rand.Intn(100) < 10 {
			log.Println(" ¡Trevor se emborrachó! Misión de distracción fracasada.")
			s.missionFailed = true
			return
		}
	}

	if !s.missionFailed {
		s.missionSuccess = true
		log.Println("Trevor completó la distracción con éxito!")
	}
}

func (s *trevorServer) workOnGolpe() {
	for s.currentTurns < s.totalTurns && !s.missionFailed {
		time.Sleep(10 * time.Millisecond)
		s.currentTurns++
	}

	if !s.missionFailed {
		s.missionSuccess = true
		s.finalLoot = s.baseLoot  
		log.Printf("Trevor completó el golpe con éxito! Botín final: $%d", s.finalLoot)
	}
}

func (s *trevorServer) GetFinalLoot(ctx context.Context, req *pb.LootRequest) (*pb.LootResponse, error) {
	if s.missionSuccess {
		log.Printf("Trevor entregando botín final: $%d", s.finalLoot)
		return &pb.LootResponse{FinalLoot: s.finalLoot}, nil
	}
	return nil, fmt.Errorf("la misión de Trevor no ha sido completada con éxito")
}

func (s *trevorServer) CheckStatus(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	status := "waiting"
	extraLoot := int32(0)

	if s.isWorking {
		status = "working"
	}
	if s.missionSuccess {
		status = "success"
		extraLoot = s.finalLoot - s.baseLoot
	}
	if s.missionFailed {
		status = "failed"
	}

	return &pb.StatusResponse{
		Status:         status,
		TurnsCompleted: s.currentTurns,
		TotalTurns:     s.totalTurns,
		CurrentStars:   s.currentStars,
		ExtraLoot:      extraLoot,
	}, nil
}

func (s *trevorServer) ReceivePayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
    log.Printf("Trevor recibió pago de $%d", req.Amount)

    if s.finalLoot > 0 {
        expectedAmount := s.finalLoot / 4
        if req.Amount == expectedAmount {
            return &pb.PaymentResponse{
                Message: "¡Justo lo que esperaba!",
                CorrectAmount: true,
            }, nil
        }
        return &pb.PaymentResponse{
            Message: fmt.Sprintf("Error: Esperaba $%d, recibí $%d", expectedAmount, req.Amount),
            CorrectAmount: false,
        }, nil
    }
    
    return &pb.PaymentResponse{
        Message: "¡Justo lo que esperaba!",
        CorrectAmount: true,
    }, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50063")
	if err != nil {
		log.Fatalf("Error al escuchar: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMissionServiceServer(grpcServer, &trevorServer{})

	log.Println("Servidor de Trevor escuchando en :50063")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error en gRPC: %v", err)
	}
}
