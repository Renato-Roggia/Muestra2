//lester

package main

import (
	"context"
	"encoding/csv"
	"math/rand"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	pb "Tarea/proto"

	"github.com/streadway/amqp"

	"google.golang.org/grpc"
)

type Offer struct {
	Loot            int32
	SuccessFranklin int32
	SuccessTrevor   int32
	PoliceRisk      int32
}

type ClientState struct {
	currentOffer  int
	rejectedCount int
}

type lesterServer struct {
	pb.UnimplementedLesterServiceServer
	pb.UnimplementedNotificationServiceServer
	offers        []Offer
	rabbitConn    *amqp.Connection
	rabbitChannel *amqp.Channel
	activeStars   map[string]bool
	clientStates  map[string]*ClientState
}

func connectRabbitMQ() (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Error conectando a RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error abriendo canal: %v", err)
	}

	return conn, ch
}

func loadOffers(filename string) []Offer {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("No se pudo abrir el CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error leyendo CSV: %v", err)
	}

	var offers []Offer
	for i, row := range records {
		if i == 0 {
			continue
		}

		for j, cell := range row {
			if cell == "" {
				log.Printf("Advertencia: Fila %d, columna %d está vacía", i+1, j+1)
				continue
			}
		}

		loot, err1 := strconv.Atoi(row[0])
		sf, err2 := strconv.Atoi(row[1])
		st, err3 := strconv.Atoi(row[2])
		risk, err4 := strconv.Atoi(row[3])

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			log.Printf("Advertencia: Fila %d tiene valores inválidos", i+1)
			continue
		}

		if loot < 0 || sf < 0 || sf > 100 || st < 0 || st > 100 || risk < 0 || risk > 100 {
			log.Printf("Advertencia: Fila %d tiene valores fuera de rango", i+1)
			continue
		}

		offers = append(offers, Offer{
			Loot:            int32(loot),
			SuccessFranklin: int32(sf),
			SuccessTrevor:   int32(st),
			PoliceRisk:      int32(risk),
		})
	}
	return offers
}

func (s *lesterServer) GetOffer(ctx context.Context, req *pb.OfferRequest) (*pb.OfferResponse, error) {
	// Posibilidad de que no tenga ofertas
	if rand.Intn(100) >= 90 {
		log.Printf("Lester no tiene trabajo disponible (10%% probabilidad)")
		return &pb.OfferResponse{HasOffer: false}, nil
	}
	
	// Inicializar estado del cliente si no existe
	if _, exists := s.clientStates[req.Requester]; !exists {
		s.clientStates[req.Requester] = &ClientState{
			currentOffer:  0, // Siempre comenzar desde 0 para nuevo cliente
			rejectedCount: 0,
		}
		log.Printf("Nuevo cliente registrado: %s", req.Requester)
	}

	clientState := s.clientStates[req.Requester]

	if clientState.currentOffer >= len(s.offers) {
		log.Printf("No hay más ofertas válidas para %s", req.Requester)
		return &pb.OfferResponse{HasOffer: false}, nil
	}

	if clientState.rejectedCount >= 3 {
		log.Printf("%s rechazó 3 veces. Esperando 10 segundos...", req.Requester)
		time.Sleep(10 * time.Second)
		clientState.rejectedCount = 0
	}

	offer := s.offers[clientState.currentOffer]
	log.Printf("Ofreciendo oferta %d/%d a %s: Botín=%d, F=%d%%, T=%d%%, Riesgo=%d%%",
		clientState.currentOffer+1, len(s.offers), req.Requester, offer.Loot,
		offer.SuccessFranklin, offer.SuccessTrevor, offer.PoliceRisk)

	return &pb.OfferResponse{
		HasOffer:        true,
		Loot:            offer.Loot,
		SuccessFranklin: offer.SuccessFranklin,
		SuccessTrevor:   offer.SuccessTrevor,
		PoliceRisk:      offer.PoliceRisk,
	}, nil
}


func (s *lesterServer) ConfirmDecision(ctx context.Context, req *pb.DecisionRequest) (*pb.DecisionResponse, error) {
	clientState := s.clientStates[req.Requester]

	if req.Accepted {
		log.Printf("%s aceptó la oferta %d", req.Requester, clientState.currentOffer+1)
		clientState.rejectedCount = 0
		clientState.currentOffer++
		return &pb.DecisionResponse{Message: "Perfecto, comenzamos el atraco."}, nil
	}

	clientState.rejectedCount++
	log.Printf("%s rechazó la oferta %d (rechazos consecutivos: %d)",
		req.Requester, clientState.currentOffer+1, clientState.rejectedCount)

	clientState.currentOffer++ // Avanzar a siguiente oferta
	return &pb.DecisionResponse{Message: "Ok, buscaré otra opción..."}, nil
}

func (s *lesterServer) StartStarNotifications(ctx context.Context, req *pb.StarRequest) (*pb.StarResponse, error) {
	log.Printf("Iniciando notificaciones de estrellas para %s", req.Character)
	s.activeStars[req.Character] = true

	go s.sendStarNotifications(req.Character, req.PoliceRisk)

	return &pb.StarResponse{Success: true}, nil
}

func (s *lesterServer) sendStarNotifications(character string, policeRisk int32) {
	frequency := 100 - policeRisk
	if frequency < 10 {
		frequency = 10
	}

	ticker := time.NewTicker(time.Duration(frequency) * 100 * time.Millisecond)
	defer ticker.Stop()

	stars := 0
	for s.activeStars[character] && stars <= 7 {
		<-ticker.C
		stars++

		err := s.rabbitChannel.Publish(
			"",
			"stars_"+character,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(strconv.Itoa(stars)),
			})
		if err != nil {
			log.Printf("Error publicando estrella: %v", err)
		}

		log.Printf(" Estrella %d enviada a %s", stars, character)
	}
}

func (s *lesterServer) StopStarNotifications(ctx context.Context, req *pb.StopRequest) (*pb.StopResponse, error) {
	log.Printf("Deteniendo notificaciones para %s", req.Character)
	s.activeStars[req.Character] = false
	return &pb.StopResponse{Success: true}, nil
}

func (s *lesterServer) ReceivePayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
    log.Printf("Lester recibió pago de $%d", req.Amount)
	
    if req.Amount > 0 {
        return &pb.PaymentResponse{
            Message: "Un placer hacer negocios.",
            CorrectAmount: true,
        }, nil
    }
    
    return &pb.PaymentResponse{
        Message: "El pago no es correcto.",
        CorrectAmount: false,
    }, nil
}

func (s *lesterServer) SendFinalReport(ctx context.Context, req *pb.FinalReport) (*pb.ReportResponse, error) {
    log.Printf("Reporte final de la misión recibido:")
    log.Printf("  Estado: %s", req.MissionOutcome)
    log.Printf("  Botín Total: $%d", req.TotalLoot)
    log.Printf("  Reparto: Michael $%d, Franklin $%d, Trevor $%d, Lester $%d",
        req.MichaelShare, req.FranklinShare, req.TrevorShare, req.LesterShare)

    if req.MissionOutcome == "failed" {
        log.Printf("  La misión fracasó debido a: %s", req.ErrorMessage)
    }

    return &pb.ReportResponse{Message: "Reporte recibido y procesado."}, nil
}

func main() {
	rabbitConn, rabbitChannel := connectRabbitMQ()
	defer rabbitConn.Close()
	defer rabbitChannel.Close()

	offers := loadOffers("ofertas.csv")
	lis, err := net.Listen("tcp", ":50061")
	if err != nil {
		log.Fatalf("Error al escuchar: %v", err)
	}

	grpcServer := grpc.NewServer()
	server := &lesterServer{
		offers:        offers,
		rabbitConn:    rabbitConn,
		rabbitChannel: rabbitChannel,
		activeStars:   make(map[string]bool),
		clientStates:  make(map[string]*ClientState),
	}

	pb.RegisterLesterServiceServer(grpcServer, server)
	pb.RegisterNotificationServiceServer(grpcServer, server)

	log.Println("Servidor de Lester escuchando en :50061")
	log.Printf("Cargadas %d ofertas desde CSV", len(offers))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error en gRPC: %v", err)
	}
}
