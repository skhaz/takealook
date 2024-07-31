package router

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentlink"
	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
)

func init() {
	stripe.Key = os.Getenv("STRIPE_APIKEY")
}

func (r *Router) Pay(c echo.Context) error {
	result, err := paymentlink.New(&stripe.PaymentLinkParams{
		AfterCompletion: &stripe.PaymentLinkAfterCompletionParams{
			Type: stripe.String("redirect"),
			Redirect: &stripe.PaymentLinkAfterCompletionRedirectParams{
				URL: stripe.String(fmt.Sprintf("%s/dashboard", domain)),
			},
		},
		LineItems: []*stripe.PaymentLinkLineItemParams{
			{
				Price:    stripe.String(os.Getenv("STRIPE_PRICE_ID")),
				Quantity: stripe.Int64(1),
			},
		},
	})
	if err != nil {
		var message = "error while creating payment link"
		log.Error(message, zap.Error(err))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": message})
	}

	return c.Render(http.StatusOK, "pay", struct{ Link string }{Link: result.URL})
}
