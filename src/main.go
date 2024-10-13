package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/paymentintent"
)

func main() {
	stripe.Key = "sk_test_51PiHVgRpvZCfjfCQHqnXh0gjVETqfKzBSx4xsEPNR9fwl2JT4GvMxOP9g9aVRZXrSrA7UNJx3lVFaUzXQzzFCjCc006tOkf9k4"
	http.HandleFunc("/create-payment-intent", handleCreatePaymentIntent)
	http.HandleFunc("/health", handleHealth)

	log.Println("Server is running")
	err := http.ListenAndServe("localhost:4242", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleCreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProductId string `json:"product_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Address1  string `json:"address_1"`
		Address2  string `json:"address_2"`
		City      string `json:"city"`
		State     string `json:"state"`
		Zip       string `json:"zip"`
		Country   string `json:"country"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	amount := calculateOrderAmount(req.ProductId)
	if amount < 50 { // Minimum amount in cents
		http.Error(w, "Amount must be at least 50 cents", http.StatusBadRequest)
		return
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	paymentIntent, err := paymentintent.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response struct {
		ClientSecret string `json:"clientSecret"`
	}

	response.ClientSecret = paymentIntent.ClientSecret

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = io.Copy(w, &buf)
	if err != nil {
		fmt.Println(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := []byte("Server is up and Running")
	_, err := w.Write(resp)
	if err != nil {
		fmt.Println(err)
	}
}

func calculateOrderAmount(ProductId string) int64 {
	switch ProductId {
	case "Forever Pants":
		return 26000 // $26.00
	case "Forever Shirts":
		return 15500 // $15.50
	case "Forever Shorts":
		return 30000 // $30.00
	}
	return 0
}
