import { useState, useEffect } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { Package, CheckCircle, Clock, Truck, XCircle } from 'lucide-react';
import { orderApi } from '../api/orders';
import type { Order, OrderStatus } from '../types';
import Loading from '../components/Loading';

const statusConfig: Record<OrderStatus, { icon: typeof Package; color: string; bg: string }> = {
  pending: { icon: Clock, color: 'text-yellow-600', bg: 'bg-yellow-100' },
  confirmed: { icon: CheckCircle, color: 'text-blue-600', bg: 'bg-blue-100' },
  processing: { icon: Package, color: 'text-indigo-600', bg: 'bg-indigo-100' },
  shipped: { icon: Truck, color: 'text-purple-600', bg: 'bg-purple-100' },
  delivered: { icon: CheckCircle, color: 'text-green-600', bg: 'bg-green-100' },
  cancelled: { icon: XCircle, color: 'text-red-600', bg: 'bg-red-100' },
};

export default function Orders() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchParams] = useSearchParams();
  const showSuccess = searchParams.get('success') === 'true';

  useEffect(() => {
    const fetchOrders = async () => {
      try {
        const data = await orderApi.getAll(1, 20);
        setOrders(data.orders || []);
      } catch (err) {
        console.error('Failed to fetch orders:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchOrders();
  }, []);

  if (loading) return <Loading />;

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">My Orders</h1>

      {showSuccess && (
        <div className="mb-6 p-4 bg-green-50 text-green-700 rounded-lg flex items-center space-x-2">
          <CheckCircle className="h-5 w-5" />
          <span>Your order has been placed successfully!</span>
        </div>
      )}

      {orders.length === 0 ? (
        <div className="text-center py-12">
          <Package className="h-16 w-16 text-gray-400 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-gray-900 mb-2">No orders yet</h2>
          <p className="text-gray-600 mb-8">Start shopping to see your orders here</p>
          <Link
            to="/products"
            className="inline-flex items-center space-x-2 bg-indigo-600 text-white px-6 py-3 rounded-lg hover:bg-indigo-700"
          >
            <span>Browse Products</span>
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {orders.map((order) => {
            const status = statusConfig[order.status] || statusConfig.pending;
            const StatusIcon = status.icon;
            
            return (
              <div key={order.orderId} className="bg-white rounded-lg shadow-md p-6">
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <p className="text-sm text-gray-500">Order ID</p>
                    <p className="font-mono text-sm">{order.orderId}</p>
                  </div>
                  <div className={`flex items-center space-x-2 px-3 py-1 rounded-full ${status.bg}`}>
                    <StatusIcon className={`h-4 w-4 ${status.color}`} />
                    <span className={`text-sm font-medium capitalize ${status.color}`}>
                      {order.status}
                    </span>
                  </div>
                </div>

                <div className="border-t pt-4">
                  <div className="space-y-2">
                    {order.items.map((item, idx) => (
                      <div key={idx} className="flex justify-between text-sm">
                        <span className="text-gray-600">
                          {item.name} x {item.quantity}
                        </span>
                        <span>${item.subtotal.toFixed(2)}</span>
                      </div>
                    ))}
                  </div>
                  <hr className="my-3" />
                  <div className="flex justify-between font-bold">
                    <span>Total</span>
                    <span>${order.totalAmount.toFixed(2)}</span>
                  </div>
                </div>

                <div className="mt-4 pt-4 border-t">
                  <p className="text-sm text-gray-500">Shipping Address</p>
                  <p className="text-sm">{order.shippingAddress}</p>
                </div>

                <div className="mt-2 text-xs text-gray-400">
                  Ordered on {new Date(order.createdAt).toLocaleDateString()}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
