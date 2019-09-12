package serve

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func getMongoDB() (*mongo.Database, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://192.168.6.200:27017/test-bootstrap"
	}
	cs, err := connstring.Parse(uri)
	if err != nil {
		panic(err)
	}

	client, err := mongo.NewClient(
		options.Client().ApplyURI(uri),
	)
	client.Connect(nil)

	return client.Database(cs.Database), err
}

func getKeyPair() (pri *rsa.PrivateKey, err error) {
	block, _ := pem.Decode([]byte(pemString))
	pri, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	pri.Precompute()
	return
}

func TestEncryptDecrypt(t *testing.T) {
	pri, err := getKeyPair()
	if err != nil {
		t.Log(err)
	}

	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, &pri.PublicKey, []byte("hello"))
	if err != nil {
		t.Log(err)
	}

	p, err := pri.Decrypt(rand.Reader, cipher, nil)

	if string(p) != "hello" {
		t.Error(string(p), "hello")
	}

	txt := base64.StdEncoding.EncodeToString(cipher)

	cipher, err = base64.StdEncoding.DecodeString(txt)

	if err != nil {
		t.Log(err)
	}
	p, err = pri.Decrypt(rand.Reader, cipher, nil)

	if string(p) != "hello" {
		t.Error(string(p), "hello")
	}
}

// TestGenerateKey 使用错误的 Ticket
func TestGenerateKey(t *testing.T) {
	pri, err := getKeyPair()
	if err != nil {
		t.Log(err)
	}

	db, err := getMongoDB()
	if err != nil {
		t.Log(err)
	}

	ticket := "TestGenerateKey"
	workerID := "hello"
	_, err = db.Collection(collectionTickt).InsertOne(context.Background(),
		bootstrapTicket{ID: ticket, RemainCount: 1},
	)
	if err != nil {
		t.Log(err)
	}
	defer db.Collection(collectionTickt).DeleteOne(context.Background(), bson.M{"_id": ticket})
	// .Collection("boorstrapTickt")
	backend := BootStrapBackend{
		DB:        db,
		PublicKey: &pri.PublicKey,
	}

	gReq := generateKeyRequest{
		WorkerID: workerID,
		Ticket:   ticket,
	}
	b, err := json.Marshal(gReq)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()
	r := http.Request{
		Method: http.MethodPost,
		Body:   ioutil.NopCloser(bytes.NewReader(b)),
	}
	backend.GenerateKey(w, &r)
	defer db.Collection(collectionKey).DeleteOne(context.Background(), bson.M{"_id": gReq.WorkerID})

	if w.Code != http.StatusOK {
		t.Error(w)
	}

	var ret APIResult
	err = json.Unmarshal(w.Body.Bytes(), &ret)
	if err != nil {
		t.Error(err)
	}

	err = pri.Validate()
	if err != nil {
		t.Error(err)
	}

	key := (*ret.Data)["key"]
	bb, err := base64.StdEncoding.DecodeString(key.(string))
	if err != nil {
		t.Error(err)
	}
	p, err := pri.Decrypt(rand.Reader, bb, nil)
	if err != nil {
		t.Error(err)
	}
	if string(p) != workerID {
		t.Error("Encrypt/Decrypt Not match", string(p), workerID)
	}
	// t.Log(w.Body.String())

	// 相同WorkerID重复请求
	w = httptest.NewRecorder()
	r = http.Request{
		Method: http.MethodPost,
		Body:   ioutil.NopCloser(bytes.NewReader(b)),
	}
	backend.GenerateKey(w, &r)

	if w.Code != http.StatusOK {
		t.Error(w)
	}
}
func TestGenerateKeyWithWrongTicket(t *testing.T) {
	pri, err := getKeyPair()
	db, err := getMongoDB()
	if err != nil {
		t.Log(err)
	}
	// .Collection("boorstrapTickt")
	backend := BootStrapBackend{
		DB: db,
		// PrivateKey: pri,
		PublicKey: &pri.PublicKey,
	}

	gReq := generateKeyRequest{
		WorkerID: "hello",
		Ticket:   "ha",
	}
	b, err := json.Marshal(gReq)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()
	r := http.Request{
		Method: http.MethodPost,
		Body:   ioutil.NopCloser(bytes.NewReader(b)),
	}
	backend.GenerateKey(w, &r)

	if w.Code != http.StatusUnauthorized {
		t.Error(w)
	}
}
