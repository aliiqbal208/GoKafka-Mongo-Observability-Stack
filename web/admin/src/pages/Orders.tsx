import { useEffect, useState } from 'react';
import { RefreshCw } from 'lucide-react';
import Loading from '../components/Loading';
import OrderTable from '../components/OrderTable';
import { getOrders, updateOrderStatus } from '../api/orders';
import { Order, OrderStatus } from '../types';

const statusFilters: { label: string; value: OrderStatus | 'all' }[] = [
  { label: 'All', value: 'all' },
  { label: 'Pending', value: 'pending' },
  { label: 'Processing', value: 'processing' },
  { label: 'Shipped', value: 'shipped' },
  { label: 'Delivered', value: 'delivered' },
  { label: 'Cancelled', value: 'cancelled' },
];

export default function Orders() {
  const [loading, setLoading] = useState(true);
  const [orders, setOrders] = useState<Order[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [filter, setFilter] = useState<OrderStatus | 'all'>('all');
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);

  const fetchOrders = async () => {
    setLoading(true);
    try {
      const res = await getOrders(page, 10);
      let ordersList = res.orders || [];
      
      if (filter !== 'all') {
        ordersList = ordersList.filter((o: Order) => o.status === filter);
      }
      
      setOrders(ordersList);
      setTotalPages(res.totalPages || 1);
    } catch (error) {
      console.error('Failed to fetch orders:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOrders();
  }, [page, filter]);

  const handleStatusChange = async (orderId: string, newStatus: OrderStatus) => {
    try {
      await updateOrderStatus(orderId, newStatus);
      fetchOrders();
      setSelectedOrder(null);
    } catch (error) {
      console.error('Failed to update order status:', error);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Orders</h1>
          <p className="text-gray-500">Manage and track customer orders</p>
        </div>
        <button
          onClick={fetchOrders}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
        >
          <RefreshCw className="w-5 h-5" />
          Refresh
        </button>
      </div>

      {/* Status Filters */}
      <div className="flex gap-2 flex-wrap">
        {statusFilters.map((s) => (
          <button
            key={s.value}
            onClick={() => setFilter(s.value)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              filter === s.value
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
            }`}
          >
            {s.label}
          </button>
        ))}
      </div>

      {loading ? (
        <Loading />
      ) : (
        <>
          <OrderTable orders={orders} onStatusChange={handleStatusChange} />

          {/* Pagination */}
          <div className="flex items-center justify-between">
            <p className="text-sm text-gray-500">Page {page} of {totalPages}</p>
            <div className="flex gap-2">
              <button
                onClick={() => setPage(p => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-3 py-1 border rounded-lg disabled:opacity-50"
              >
                Previous
              </button>
              <button
                onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="px-3 py-1 border rounded-lg disabled:opacity-50"
              >
                Next
              </button>
            </div>
          </div>
        </>
      )}

      {/* Order Detail Modal */}
      {selectedOrder && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-xl p-6 w-full max-w-lg">
            <h2 className="text-xl font-bold mb-4">Order #{selectedOrder.orderId.slice(-8)}</h2>
            
            <div className="space-y-4">
              <div>
                <p className="text-sm text-gray-500">Customer</p>
                <p className="font-medium">{selectedOrder.userId}</p>
              </div>
              
              <div>
                <p className="text-sm text-gray-500">Items</p>
                <ul className="mt-2 space-y-2">
                  {selectedOrder.items?.map((item, idx) => (
                    <li key={idx} className="flex justify-between">
                      <span>{item.productName} x{item.quantity}</span>
                      <span>${(item.price * item.quantity).toFixed(2)}</span>
                    </li>
                  ))}
                </ul>
              </div>
              
              <div className="border-t pt-4">
                <div className="flex justify-between font-bold">
                  <span>Total</span>
                  <span>${selectedOrder.totalAmount?.toFixed(2)}</span>
                </div>
              </div>

              <div>
                <p className="text-sm text-gray-500 mb-2">Update Status</p>
                <select
                  value={selectedOrder.status}
                  onChange={(e) => handleStatusChange(selectedOrder.orderId, e.target.value as OrderStatus)}
                  className="w-full border rounded-lg px-3 py-2"
                >
                  <option value="pending">Pending</option>
                  <option value="processing">Processing</option>
                  <option value="shipped">Shipped</option>
                  <option value="delivered">Delivered</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </div>
            </div>

            <button
              onClick={() => setSelectedOrder(null)}
              className="mt-6 w-full px-4 py-2 border rounded-lg hover:bg-gray-50"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
