package v1

import "github.com/prometheus/client_golang/prometheus"

var (
	getCartRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cart_get_requests_total",
		Help: "Total number of get cart requests",
	})
	addItemRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cart_add_item_requests_total",
		Help: "Total number of add item requests",
	})
	updateItemRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cart_update_item_requests_total",
		Help: "Total number of update item requests",
	})
	removeItemRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cart_remove_item_requests_total",
		Help: "Total number of remove item requests",
	})
	clearCartRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cart_clear_requests_total",
		Help: "Total number of clear cart requests",
	})
	successRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cart_success_requests_total",
		Help: "Total number of successful cart requests",
	})
	errorRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cart_error_requests_total",
		Help: "Total number of failed cart requests",
	})
)

func init() {
	prometheus.MustRegister(
		getCartRequests,
		addItemRequests,
		updateItemRequests,
		removeItemRequests,
		clearCartRequests,
		successRequests,
		errorRequests,
	)
}
