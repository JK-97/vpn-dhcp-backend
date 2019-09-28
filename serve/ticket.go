package serve

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"dhcp-backend/go-utils/logger"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	collectionTickt = "boorstrapTickt"
	collectionKey   = "boorstrapKey"
)

// BootStrapBackend BootStrap 后端
type BootStrapBackend struct {
	// PrivateKey *rsa.PrivateKey // 私钥
	PublicKey *rsa.PublicKey // 公钥

	DB *mongo.Database
}

type generateKeyRequest struct {
	WorkerID string `json:"wid"`
	Ticket   string `json:"ticket"`
}

type bootstrapTicket struct {
	ID          string    `bson:"_id"`
	RemainCount int       `bson:"remainCount"`
	WorkerID    string    `bson:"wid,omitempty"`
	DeadLine    time.Time `bson:"deadLine"` // 失效时间
}

type bootstrapKey struct {
	ID string `bson:"_id"`

	Key    []byte    `bson:"key"`
	Ticket string    `bson:"ticket"`
	Time   time.Time `bson:"time"`
}

func (b *BootStrapBackend) getKey(workerID string) []byte {
	c2 := b.DB.Collection(collectionKey)
	result := c2.FindOne(nil, bson.M{"_id": workerID})

	if result == nil {
		return nil
	} else if result.Err() == mongo.ErrNoDocuments {
		return nil
	}

	var key bootstrapKey
	err := result.Decode(&key)
	if err != nil {
		return nil
	}
	return key.Key
}

// generateKey 由
func generateKey(pub *rsa.PublicKey, workerID string) (p []byte) {
	p, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(workerID))
	if err != nil {
		logger.Info(err)
		return nil
	}
	return
}

func (b *BootStrapBackend) generateKey(ctx context.Context, w http.ResponseWriter, workerID string, ticket *bootstrapTicket) (p []byte) {
	collection := b.DB.Collection(collectionTickt)
	result := collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": ticket.ID, "remainCount": bson.M{"$gt": 0}},
		bson.M{"$inc": bson.M{"remainCount": -1}},
	)

	if result == nil {
		ErrorWithCode(w, http.StatusUnauthorized)
		return
	} else if result.Err() != nil {
		return
	}
	// var t bootstrapTicket
	err := result.Decode(ticket)
	if err != nil {
		ErrorWithCode(w, http.StatusInternalServerError)
		return
	}
	if ticket.WorkerID != "" && ticket.WorkerID != workerID {
		// 使用针对固定设备的 ticket 注册
		return nil
	}

	c2 := b.DB.Collection(collectionKey)

	key := bootstrapKey{
		ID:     workerID,
		Key:    generateKey(b.PublicKey, workerID),
		Ticket: ticket.ID,
		Time:   time.Now().UTC(),
	}

	_, err = c2.InsertOne(nil, key)

	if err != nil {
		// 回滚
		collection.FindOneAndUpdate(
			nil,
			bson.M{"_id": ticket.ID},
			bson.M{"$inc": bson.M{"remainCount": 1}},
		)
		return
	}
	return key.Key
}

// GenerateKey 根据 worker ID 生成对应的 Key
func (b *BootStrapBackend) GenerateKey(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)

	if err != nil {
		Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req generateKeyRequest
	err = json.Unmarshal(buf, &req)
	if err != nil {
		Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t, workerID := req.Ticket, req.WorkerID
	ticket := bootstrapTicket{
		ID: t,
	}

	keyBytes := b.getKey(workerID)
	if keyBytes == nil {
		keyBytes = b.generateKey(r.Context(), w, workerID, &ticket)

		if keyBytes == nil {
			ErrorWithCode(w, http.StatusUnauthorized)
			return
		}
	}

	keyString := base64.StdEncoding.EncodeToString(keyBytes)

	data := map[string]interface{}{
		"key":         keyString,
		"remainCount": ticket.RemainCount - 1,
	}
	if !ticket.DeadLine.IsZero() {
		data["deadLine"] = ticket.DeadLine.Unix()
	}
	logger.Info("Generate Ticket:", ticket, keyString)
	WriteData(w, &data)
}

type genTicketRequest struct {
	WorkerID    string `json:"wid"`
	TTL         int64  `json:"ttl"`
	RemainCount uint   `bson:"remainCount"`
}

// AddTicket 添加 ticket
func (b *BootStrapBackend) AddTicket(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req genTicketRequest
	err = json.Unmarshal(buf, &req)
	if err != nil {
		Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var ticket bootstrapTicket

	if req.WorkerID == "" {
		if req.RemainCount == 0 {
			Error(w, "RemainCount should grater than 0", http.StatusBadRequest)
			return
		}
		ticket.RemainCount = int(req.RemainCount)
	} else {
		if req.TTL < 86400 {
			req.TTL = 86400
		}
		ticket.WorkerID = req.WorkerID
		ticket.RemainCount = 999
		ticket.DeadLine = time.Now().Add(time.Duration(req.TTL) * time.Second)
	}

	collection := b.DB.Collection(collectionTickt)

	for {
		ticket.ID = getRandomString(6)
		_, err := collection.InsertOne(r.Context(), ticket)
		if err == nil {
			break
		}
	}

	data := map[string]interface{}{
		"Ticket":      ticket.ID,
		"RemainCount": ticket.RemainCount,
		"WorkerID":    ticket.WorkerID,
		"DeadLine":    ticket.DeadLine,
	}

	WriteData(w, &data)
}
