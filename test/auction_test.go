package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fullcycle-auction_go/configuration/database/mongodb"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/api/web/controller/auction_controller"
	"fullcycle-auction_go/internal/infra/database/auction"
	"fullcycle-auction_go/internal/infra/database/bid"
	"fullcycle-auction_go/internal/usecase/auction_usecase"
	"fullcycle-auction_go/pkg/timer"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func clearDatabase(t *testing.T, ctx context.Context) {
	database, err := mongodb.NewMongoDBConnection(ctx)
	if err != nil {
		t.Fatalf("Erro ao conectar ao MongoDB: %v", err)
		return
	}

	client := database.Client()
	defer client.Disconnect(context.Background())

	dbName := database.Name()
	t.Logf("Limpando banco: %s", dbName)

	collections, _ := client.Database(dbName).ListCollectionNames(ctx, bson.D{})
	t.Logf("Coleções disponíveis: %v", collections)

	collection := client.Database(dbName).Collection("auctions")

	res, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Erro ao limpar a coleção: %v", err)
	}
	t.Logf("Total de documentos removidos: %d", res.DeletedCount)
}

func TestAuction(t *testing.T) {

	ctx := context.Background()

	clearDatabase(t, ctx)
	t.Cleanup(func() { clearDatabase(t, ctx) })

	category := "test, product"
	productName := "test"

	duration, err := timer.AuctionTimer()
	if err != nil {
		t.Errorf("error to get auction duration. Error: %s", err)
		return
	}

	database, err := mongodb.NewMongoDBConnection(ctx)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	auctionRepository := auction.NewAuctionRepository(database)
	bidRepository := bid.NewBidRepository(database, auctionRepository)

	controller := auction_controller.NewAuctionController(auction_usecase.NewAuctionUseCase(auctionRepository, bidRepository))
	if err != nil {
		t.Errorf("error to create controller. Error: %s", err)
		return
	}

	//
	//abertura da auction
	//

	product := auction_usecase.AuctionInputDTO{
		ProductName: productName,
		Category:    category,
		Description: "test product",
		Condition:   0,
	}

	body, err := json.Marshal(product)
	if err != nil {
		t.Errorf("error to parse create auction body request. Error: %s", err)
		return
	}

	req := httptest.NewRequest(http.MethodPost, "/auction", bytes.NewReader(body))
	req.Header.Add("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	ginCtx, _ := gin.CreateTestContext(rec)

	ginCtx.Request = req

	controller.CreateAuction(ginCtx)

	assert.Equal(t, http.StatusOK, rec.Code)

	t.Log("auction created successfully")

	//
	//validando abertura da auction
	//

	req = httptest.NewRequest(http.MethodGet, "/auction", nil)
	q := req.URL.Query()
	q.Add("status", "0")
	q.Add("category", category)
	q.Add("productName", productName)
	req.URL.RawQuery = q.Encode()

	rec = httptest.NewRecorder()

	ginCtx, _ = gin.CreateTestContext(rec)

	ginCtx.Request = req

	controller.FindAuctions(ginCtx)

	if !assert.Equal(t, http.StatusOK, rec.Code) {
		return
	}

	var res []auction_entity.Auction

	err = json.NewDecoder(rec.Body).Decode(&res)
	if err != nil {
		t.Errorf("error to parse get auction response. Error: %s", err)
	}

	if len(res) == 0 {
		t.Error("no auction was found")
	}

	for _, auction := range res {
		assert.Equal(t, auction.Status, auction_entity.Active)
	}

	t.Log("auction validated successfully")

	//
	//validando fechamento da auction
	//

	time.Sleep(duration)

	req = httptest.NewRequest(http.MethodGet, "/auction", nil)
	q = req.URL.Query()
	q.Add("status", "1")
	q.Add("category", category)
	q.Add("productName", productName)
	req.URL.RawQuery = q.Encode()

	rec = httptest.NewRecorder()

	ginCtx, _ = gin.CreateTestContext(rec)

	ginCtx.Request = req

	controller.FindAuctions(ginCtx)

	if !assert.Equal(t, http.StatusOK, rec.Code) {
		return
	}

	err = json.NewDecoder(rec.Body).Decode(&res)
	if err != nil {
		t.Errorf("error to parse get auction response. Error: %s", err)
	}

	if len(res) == 0 {
		t.Error("no auction was found")
	}

	for _, auction := range res {
		if assert.Equal(t, auction_entity.Completed, auction.Status) {
			t.Log("auction closed successfuly")
		} else {
			t.Log("auction not closed")
		}
	}

}
