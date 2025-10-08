
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

type franklinServer struct {
	pb.UnimplementedMissionServiceServer
	currentTurns   int32
	totalTurns     int32
	isWorking      bool
	missionFailed  bool
	missionSuccess bool
	currentStars   int32
	extraLoot      int32
	abilityActive  bool
	baseLoot       int32
	finalLoot      int32
}

func (s *franklinServer) StartDistraction(ctx context.Context, req *pb.DistractionRequest) (*pb.DistractionResponse, error) {
	s.totalTurns = req.RequiredTurns
	s.currentTurns = 0
	s.isWorking = true
	s.missionFailed = false
	s.missionSuccess = false

	log.Printf("Franklin iniciando distracción. Turnos requeridos: %d", s.totalTurns)
	go s.workOnDistraction()

	return &pb.DistractionResponse{
		Success: true,
		Message: "Franklin comenzó la distracción",
	}, nil
}

func (s *franklinServer) StartGolpe(ctx context.Context, req *pb.GolpeRequest) (*pb.GolpeResponse, error) {
	s.totalTurns = req.RequiredTurns
	s.currentTurns = 0
	s.isWorking = true
	s.missionFailed = false
	s.missionSuccess = false
	s.baseLoot = req.BaseLoot  
	s.extraLoot = 0
	s.abilityActive = false
	s.currentStars = 0
	s.finalLoot = req.BaseLoot  

	log.Printf("Franklin iniciando golpe. Turnos requeridos: %d, Botín base: $%d", 
		s.totalTurns, s.baseLoot)

	// Consumir estrellas de RabbitMQ
	go s.consumeStars()
	go s.workOnGolpe()

	return &pb.GolpeResponse{
		Success: true,
		Message: "Franklin comenzó el golpe",
	}, nil
}

func (s *franklinServer) consumeStars() {
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
		"stars_Franklin",
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
		log.Printf("Franklin - Estrellas actualizadas: %d", s.currentStars)

		// Habilidad especial de Franklin
		if s.currentStars >= 3 && !s.abilityActive {
			log.Println(" ¡Chop activado! Generando $1000 extra por turno")
			s.abilityActive = true
		}

		// Verificar fracaso
		if s.currentStars >= 5 {
			log.Printf(" Demasiadas estrellas (%d)! Misión fracasada", s.currentStars)
			s.missionFailed = true
			break
		}
	}
}

func (s *franklinServer) workOnDistraction() {
	for s.currentTurns < s.totalTurns && !s.missionFailed {
		time.Sleep(10 * time.Millisecond)
		s.currentTurns++

		if s.currentTurns == s.totalTurns/2 && rand.Intn(100) < 10 {
			log.Println("¡Chop ladró! Misión de distracción fracasada.")
			s.missionFailed = true
			return
		}
	}

	if !s.missionFailed {
		s.missionSuccess = true
		log.Println("Franklin completó la distracción con éxito!")
	}
}

func (s *franklinServer) workOnGolpe() {
	for s.currentTurns < s.totalTurns && !s.missionFailed {
		time.Sleep(10 * time.Millisecond)
		s.currentTurns++

		if s.abilityActive {
			s.extraLoot += 1000
			s.finalLoot = s.baseLoot + s.extraLoot  
		}
	}

	if !s.missionFailed {
		s.missionSuccess = true
		s.finalLoot = s.baseLoot + s.extraLoot  
		log.Printf("Franklin completó el golpe con éxito! Loot extra: $%d, Botín final: $%d", 
			s.extraLoot, s.finalLoot)
	}
}

func (s *franklinServer) GetFinalLoot(ctx context.Context, req *pb.LootRequest) (*pb.LootResponse, error) {
	if s.missionSuccess {
		log.Printf("Franklin entregando botín final: $%d", s.finalLoot)
		return &pb.LootResponse{FinalLoot: s.finalLoot}, nil
	}
	return nil, fmt.Errorf("la misión de Franklin no ha sido completada con éxito")
}

func (s *franklinServer) CheckStatus(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	status := "waiting"
	if s.isWorking {
		status = "working"
	}
	if s.missionSuccess {
		status = "success"
	}
	if s.missionFailed {
		status = "failed"
	}

	return &pb.StatusResponse{
		Status:         status,
		TurnsCompleted: s.currentTurns,
		TotalTurns:     s.totalTurns,
		CurrentStars:   s.currentStars,
		ExtraLoot:      s.extraLoot,
	}, nil
}

func (s *franklinServer) ReceivePayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
    log.Printf("Franklin recibió pago de $%d", req.Amount)
    
    if s.finalLoot > 0 {
        expectedAmount := s.finalLoot / 4
        if req.Amount == expectedAmount {
            return &pb.PaymentResponse{
                Message: "¡Excelente! El pago es correcto.",
                CorrectAmount: true,
            }, nil
        }
        return &pb.PaymentResponse{
            Message: fmt.Sprintf("Error: Esperaba $%d, recibí $%d", expectedAmount, req.Amount),
            CorrectAmount: false,
        }, nil
    }
	//Pago por distraccion
    return &pb.PaymentResponse{
        Message: "¡Excelente! El pago es correcto.",
        CorrectAmount: true,
    }, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50062")
	if err != nil {
		log.Fatalf("Error al escuchar: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMissionServiceServer(grpcServer, &franklinServer{})

	log.Println("Servidor de Franklin escuchando en :50062")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error en gRPC: %v", err)
	}
}
