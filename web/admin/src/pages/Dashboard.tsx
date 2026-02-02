import { useEffect, useState } from 'react';
import { ShoppingCart, Package, Users, DollarSign, Clock } from 'lucide-react';
import StatCard from '../components/StatCard';
import OrderTable from '../components/OrderTable';
import Loading from '../components/Loading';
import { getOrders } from '../api/orders';
import { getProducts } from '../api/products';
import { Order, Product } from '../types';

export default function Dashboard() {
  const [loading, setLoading] = useState(true);
  const [orders, setOrders] = useState<Order[]>([]);
  const [stats, setStats] = useState({
    totalOrders: 0,
    totalRevenue: 0,
    totalProducts: 0,
    pendingOrders: 0,
  });

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [ordersRes, productsRes] = await Promise.all([
          getOrders(1, 100),
          getProducts(1, 1),
        ]);

        const ordersList = ordersRes.orders || [];
        setOrders(ordersList.slice(0, 5));

        const totalRevenue = ordersList.reduce((sum: number, o: Order) => sum + (o.totalAmount || 0), 0);
        const pendingOrders = ordersList.filter((o: Order) => o.status === 'pending').length;

        setStats({
          totalOrders: ordersRes.totalCount || ordersList.length,
          totalRevenue,
          totalProducts: productsRes.totalCount || 0,
          pendingOrders,
        });
      } catch (error) {
        console.error('Failed to fetch dashboard data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  if (loading) {
    return <Loading size="lg" />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-500">Welcome back! Here's what's happening.</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Orders"
          value={stats.totalOrders}
          icon={ShoppingCart}
          color="blue"
          trend={{ value: 12, isPositive: true }}
        />
        <StatCard
          title="Total Revenue"
          value={`$${stats.totalRevenue.toFixed(2)}`}
          icon={DollarSign}
          color="green"
          trend={{ value: 8, isPositive: true }}
        />
        <StatCard
          title="Total Products"
          value={stats.totalProducts}
          icon={Package}
          color="purple"
        />
        <StatCard
          title="Pending Orders"
          value={stats.pendingOrders}
          icon={Clock}
          color="yellow"
        />
      </div>

      {/* Recent Orders */}
      <div>
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Orders</h2>
        <OrderTable orders={orders} />
      </div>
    </div>
  );
}
