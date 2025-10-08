// michael

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "Tarea/proto"

	"google.golang.org/grpc"
)

func startDistractionPhase(client pb.MissionServiceClient, character string, successRate int32) bool {
	// Calcular turnos necesarios
	turnsRequired := 200 - successRate

	log.Printf("Enviando a %s a mision de distraccion (%d turnos)", character, turnsRequired)

	// Iniciar distraccion
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.StartDistraction(ctx, &pb.DistractionRequest{
		RequiredTurns:     turnsRequired,
		AssignedCharacter: character,
	})

	if err != nil {
		log.Fatalf("Error iniciando distraccion: %v", err)
		return false
	}

	// Monitorear progreso
	for {
		time.Sleep(1 * time.Second)

		statusResp, err := client.CheckStatus(ctx, &pb.StatusRequest{Character: character})
		if err != nil {
			log.Fatalf("Error consultando estado: %v", err)
			return false
		}

		log.Printf("Estado de %s: %s (%d/%d turnos)",
			character, statusResp.Status, statusResp.TurnsCompleted, statusResp.TotalTurns)

		if statusResp.Status == "success" {
			return true
		}
		if statusResp.Status == "failed" {
			return false
		}
	}
}

func startGolpePhase(missionClient pb.MissionServiceClient, notificationClient pb.NotificationServiceClient,
	character string, successRate int32, policeRisk int32, baseLoot int32) (bool, int32) {

	turnsRequired := 200 - successRate
	log.Printf("Enviando a %s a mision de golpe (%d turnos)", character, turnsRequired)

	// Iniciar notificaciones de estrellas
	ctxNotify := context.Background()
	_, err := notificationClient.StartStarNotifications(ctxNotify, &pb.StarRequest{
		Character:  character,
		PoliceRisk: policeRisk,
	})
	if err != nil {
		log.Fatalf("Error iniciando notificaciones: %v", err)
		return false, 0
	}

	// Iniciar golpe
	ctxGolpe := context.Background()
	_, err = missionClient.StartGolpe(ctxGolpe, &pb.GolpeRequest{
		RequiredTurns:     turnsRequired,
		AssignedCharacter: character,
		PoliceRisk:        policeRisk,
		BaseLoot:          baseLoot,
	})
	if err != nil {
		log.Fatalf("Error iniciando golpe: %v", err)
		return false, 0
	}

	// Monitorear progreso
	ctxMonitor := context.Background()
	var totalLoot int32 = baseLoot

	for {
		time.Sleep(1 * time.Second)

		statusResp, err := missionClient.CheckStatus(ctxMonitor, &pb.StatusRequest{Character: character})
		if err != nil {
			log.Fatalf("Error consultando estado: %v", err)
			return false, 0
		}

		log.Printf("Estado de %s: %s (%d/%d turnos, %d estrellas, +$%d)",
			character, statusResp.Status, statusResp.TurnsCompleted,
			statusResp.TotalTurns, statusResp.CurrentStars, statusResp.ExtraLoot)

		if statusResp.Status == "success" {
			totalLoot += statusResp.ExtraLoot

			// Detener notificaciones
			_, err := notificationClient.StopStarNotifications(ctxMonitor, &pb.StopRequest{Character: character})
			if err != nil {
				log.Printf("Error deteniendo notificaciones: %v", err)
			}

			return true, totalLoot
		}
		if statusResp.Status == "failed" {
			// Detener notificaciones
			notificationClient.StopStarNotifications(ctxMonitor, &pb.StopRequest{Character: character})
			return false, 0
		}
	}
}

func generateSuccessReport(baseLoot, extraLoot, totalLoot, franklinShare, trevorShare, lesterShare, lesterExtra int32, missionID int) {
	file, err := os.Create("/root/reports/Reporte.txt")
	if err != nil {
		log.Printf("Error creando reporte: %v", err)
		return
	}
	defer file.Close()

	content := fmt.Sprintf(`=========================================================
== REPORTE FINAL DE LA MISION ==
=========================================================
Mision: Asalto al Banco #%d
Resultado Global: MISION COMPLETADA CON EXITO!

--- REPARTO DEL BOTIN ---
Botin Base: $%d
Botin Extra (Habilidad de Chop): $%d
Botin Total: $%d

--------------------------------------------------------
Pago a Franklin: $%d
Pago a Trevor: $%d
Pago a Lester: $%d (reparto) + $%d (resto)
--------------------------------------------------------
Saldo Final de la Operacion: $%d
=========================================================`,
		missionID, baseLoot, extraLoot, totalLoot,
		franklinShare, trevorShare,
		lesterShare-lesterExtra, lesterExtra,
		totalLoot)

	file.WriteString(content)
	log.Println("Reporte.txt generado exitosamente")
}

func generateFailureReport(phase string, character string, lostLoot int32, reason string, missionID int) {
	file, err := os.Create("/root/reports/Reporte.txt")
	if err != nil {
		log.Printf("Error creando reporte: %v", err)
		return
	}
	defer file.Close()

	content := fmt.Sprintf(`=========================================================
== REPORTE FINAL DE LA MISION ==
=========================================================
Mision: Asalto al Banco #%d
Resultado Global: MISION FRACASADA

--- DETALLES DEL FRACASO ---
Fase de Fracaso: %s
Personaje Responsable: %s
Botin Perdido: $%d
Motivo: %s
=========================================================`,
		missionID, phase, character, lostLoot, reason)

	file.WriteString(content)
	log.Printf("Reporte de fracaso generado: Fase %s, %s fracaso por %s", phase, character, reason)
}

func main() {
	missionID := int(time.Now().Unix() % 10000)

	// FASE 1: Conexion con Lester
	lesterConn, err := grpc.Dial("localhost:50061", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar a Lester: %v", err)
	}
	defer lesterConn.Close()

	lesterClient := pb.NewLesterServiceClient(lesterConn)
	var accepted bool
	var currentOffer *pb.OfferResponse

	// Negociacion con Lester
	for !accepted {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		offer, err := lesterClient.GetOffer(ctx, &pb.OfferRequest{Requester: "Michael"})
		if err != nil {
			log.Fatalf("Error obteniendo oferta: %v", err)
		}

		if !offer.HasOffer {
			log.Println("Lester no tiene ofertas. Reintentando...")
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("Oferta: Botin=%d, Franklin=%d%%, Trevor=%d%%, Riesgo=%d%%",
			offer.Loot, offer.SuccessFranklin, offer.SuccessTrevor, offer.PoliceRisk)

		if (offer.SuccessFranklin > 50 || offer.SuccessTrevor > 50) && offer.PoliceRisk < 80 {
			accepted = true
			currentOffer = offer
			resp, _ := lesterClient.ConfirmDecision(ctx, &pb.DecisionRequest{
				Requester: "Michael",
				Accepted:  true,
			})
			log.Println("Decision:", resp.Message)
		} else {
			resp, _ := lesterClient.ConfirmDecision(ctx, &pb.DecisionRequest{
				Requester: "Michael",
				Accepted:  false,
			})
			log.Println("Decision:", resp.Message)
			time.Sleep(2 * time.Second)
		}
	}

	log.Println("Michael acepto un contrato valido.")

	// FASE 2: Distraccion
	var distractionSuccess bool
	var distractionCharacter string

	// Crear cliente de notificaciones para Fase 3
	notificationClient := pb.NewNotificationServiceClient(lesterConn)

	if currentOffer.SuccessFranklin > currentOffer.SuccessTrevor {
		// Franklin hace distraccion
		distractionCharacter = "Franklin"
		franklinConn, err := grpc.Dial("localhost:50062", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("No se pudo conectar a Franklin: %v", err)
		}
		defer franklinConn.Close()

		franklinClient := pb.NewMissionServiceClient(franklinConn)
		distractionSuccess = startDistractionPhase(franklinClient, "Franklin", currentOffer.SuccessFranklin)
	} else {
		// Trevor hace distraccion
		distractionCharacter = "Trevor"
		trevorConn, err := grpc.Dial("localhost:50063", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("No se pudo conectar a Trevor: %v", err)
		}
		defer trevorConn.Close()

		trevorClient := pb.NewMissionServiceClient(trevorConn)
		distractionSuccess = startDistractionPhase(trevorClient, "Trevor", currentOffer.SuccessTrevor)
	}

	if !distractionSuccess {
		log.Println("Fase 2 fracasada! Atraco cancelado.")
		generateFailureReport("Fase 2: Distraccion", distractionCharacter, currentOffer.Loot,
			"Imprevisto personal durante la mision", missionID)
		return
	}

	log.Println("Fase 2 completada con exito! Procediendo a Fase 3...")

	// FASE 3: Golpe
	var golpeCharacter string
	var golpeSuccessRate int32
	var golpeClient pb.MissionServiceClient

	if distractionCharacter == "Franklin" {
		// Franklin hizo distraccion, Trevor hace golpe
		golpeCharacter = "Trevor"
		golpeSuccessRate = currentOffer.SuccessTrevor
		trevorConn, err := grpc.Dial("localhost:50063", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("No se pudo conectar a Trevor: %v", err)
		}
		defer trevorConn.Close()
		golpeClient = pb.NewMissionServiceClient(trevorConn)
	} else {
		// Trevor hizo distraccion, Franklin hace golpe
		golpeCharacter = "Franklin"
		golpeSuccessRate = currentOffer.SuccessFranklin
		franklinConn, err := grpc.Dial("localhost:50062", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("No se pudo conectar a Franklin: %v", err)
		}
		defer franklinConn.Close()
		golpeClient = pb.NewMissionServiceClient(franklinConn)
	}

	golpeSuccess, totalLoot := startGolpePhase(golpeClient, notificationClient,
		golpeCharacter, golpeSuccessRate, currentOffer.PoliceRisk, currentOffer.Loot)

	if !golpeSuccess {
		log.Println("Fase 3 fracasada! Atraco cancelado.")
		generateFailureReport("Fase 3: Golpe", golpeCharacter, totalLoot,
			"Demasiadas estrellas de busqueda", missionID)
		return
	}

	log.Printf("Fase 3 completada con exito! Botin total: $%d", totalLoot)
	log.Println("Atraco completado con exito! Procediendo a reparto del botin...")

	// FASE 4: Reparto del Botin
	// Calcular partes
	ctx := context.Background()

	individualShare := totalLoot / 4
	lesterExtra := totalLoot % 4
	baseLoot := currentOffer.Loot
	extraLoot := totalLoot - baseLoot

	log.Printf("Botin total a repartir: $%d", totalLoot)
	log.Printf("Reparto del botin:")
	log.Printf("  Michael: $%d", individualShare)
	log.Printf("  Franklin: $%d", individualShare)
	log.Printf("  Trevor: $%d", individualShare)
	log.Printf("  Lester: $%d (incluyendo el extra de $%d)", individualShare+lesterExtra, lesterExtra)

	// Pagos
	var franklinResp, trevorResp, lesterResp string

	// Pagarle a Franklin
	franklinConn, err := grpc.Dial("localhost:50062", grpc.WithInsecure())
	if err == nil {
		defer franklinConn.Close()
		franklinPayClient := pb.NewMissionServiceClient(franklinConn)
		franklinPayResp, err := franklinPayClient.ReceivePayment(ctx, &pb.PaymentRequest{
			Amount: individualShare,
		})
		if err != nil {
			franklinResp = "Error en el pago"
		} else {
			franklinResp = franklinPayResp.Message
		}
	}

	// Pagarle a Trevor
	trevorConn, err := grpc.Dial("localhost:50063", grpc.WithInsecure())
	if err == nil {
		defer trevorConn.Close()
		trevorPayClient := pb.NewMissionServiceClient(trevorConn)
		trevorPayResp, err := trevorPayClient.ReceivePayment(ctx, &pb.PaymentRequest{
			Amount: individualShare,
		})
		if err != nil {
			trevorResp = "Error en el pago"
		} else {
			trevorResp = trevorPayResp.Message
		}
	}

	// Pagarle a Lester
	lesterPayResp, err := lesterClient.ReceivePayment(ctx, &pb.PaymentRequest{
		Amount: individualShare + lesterExtra,
	})
	if err != nil {
		lesterResp = "Error en el pago"
	} else {
		lesterResp = lesterPayResp.Message
	}

	log.Printf("Respuestas de pago:")
	log.Printf("  Franklin: %s", franklinResp)
	log.Printf("  Trevor: %s", trevorResp)
	log.Printf("  Lester: %s", lesterResp)

	// Generar reporte final
	generateSuccessReport(baseLoot, extraLoot, totalLoot, individualShare, individualShare,
		individualShare+lesterExtra, lesterExtra, missionID)

	// Enviar reporte final a Lester
	ctx = context.Background()
	finalReport := &pb.FinalReport{
		MissionOutcome: "success",
		TotalLoot:      totalLoot,
		MichaelShare:   individualShare,
		FranklinShare:  individualShare,
		TrevorShare:    individualShare,
		LesterShare:    individualShare + lesterExtra,
		ErrorMessage:   "",
	}

	_, err = lesterClient.SendFinalReport(ctx, finalReport)
	if err != nil {
		log.Printf("Error enviando reporte final a Lester: %v", err)
	} else {
		log.Println("Reporte final enviado con exito.")
	}

	log.Println("Mision completada exitosamente!")
}
