import { Order } from '../types';
import { Package, Clock, Truck, CheckCircle, XCircle } from 'lucide-react';

interface OrderTableProps {
  orders: Order[];
  onStatusChange?: (orderId: string, status: string) => void;
}

const statusConfig = {
  pending: { icon: Clock, color: 'text-yellow-600 bg-yellow-100', label: 'Pending' },
  processing: { icon: Package, color: 'text-blue-600 bg-blue-100', label: 'Processing' },
  shipped: { icon: Truck, color: 'text-purple-600 bg-purple-100', label: 'Shipped' },
  delivered: { icon: CheckCircle, color: 'text-green-600 bg-green-100', label: 'Delivered' },
  cancelled: { icon: XCircle, color: 'text-red-600 bg-red-100', label: 'Cancelled' },
};

export default function OrderTable({ orders, onStatusChange }: OrderTableProps) {
  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
      <table className="w-full">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Order ID
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Customer
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Items
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Total
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Status
            </th>
            <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Date
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-200">
          {orders.map((order) => {
            const status = statusConfig[order.status] || statusConfig.pending;
            const StatusIcon = status.icon;
            return (
              <tr key={order.orderId} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="text-sm font-medium text-gray-900">
                    #{order.orderId.slice(-8)}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="text-sm text-gray-600">{order.userId.slice(-8)}</span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="text-sm text-gray-600">{order.items?.length || 0} items</span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="text-sm font-medium text-gray-900">
                    ${order.totalAmount?.toFixed(2)}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${status.color}`}>
                    <StatusIcon className="w-3 h-3" />
                    {status.label}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {new Date(order.createdAt).toLocaleDateString()}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
      {orders.length === 0 && (
        <div className="text-center py-8 text-gray-500">No orders found</div>
      )}
    </div>
  );
}
