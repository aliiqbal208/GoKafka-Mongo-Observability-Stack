import { useEffect, useState } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line, PieChart, Pie, Cell } from 'recharts';
import { getOrders } from '../api/orders';
import { getProducts } from '../api/products';
import Loading from '../components/Loading';

const COLORS = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6'];

export default function Analytics() {
  const [loading, setLoading] = useState(true);
  const [ordersByStatus, setOrdersByStatus] = useState<{ name: string; value: number }[]>([]);
  const [revenueData, setRevenueData] = useState<{ date: string; revenue: number }[]>([]);
  const [topProducts, setTopProducts] = useState<{ name: string; sales: number }[]>([]);

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        const [ordersRes, productsRes] = await Promise.all([
          getOrders(1, 100),
          getProducts(1, 10),
        ]);

        const orders = ordersRes.orders || [];
        
        // Orders by status
        const statusCounts: Record<string, number> = {};
        orders.forEach((o: any) => {
          statusCounts[o.status] = (statusCounts[o.status] || 0) + 1;
        });
        setOrdersByStatus(
          Object.entries(statusCounts).map(([name, value]) => ({ name, value }))
        );

        // Mock revenue data (in real app, this would come from backend)
        const mockRevenue = [
          { date: 'Mon', revenue: 1200 },
          { date: 'Tue', revenue: 1800 },
          { date: 'Wed', revenue: 1400 },
          { date: 'Thu', revenue: 2200 },
          { date: 'Fri', revenue: 1900 },
          { date: 'Sat', revenue: 2800 },
          { date: 'Sun', revenue: 2100 },
        ];
        setRevenueData(mockRevenue);

        // Top products by rating
        const products = productsRes.products || [];
        setTopProducts(
          products
            .sort((a: any, b: any) => b.rating - a.rating)
            .slice(0, 5)
            .map((p: any) => ({ name: p.name.slice(0, 20), sales: p.rating * 10 }))
        );
      } catch (error) {
        console.error('Failed to fetch analytics:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchAnalytics();
  }, []);

  if (loading) {
    return <Loading size="lg" />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Analytics</h1>
        <p className="text-gray-500">Track your store performance</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Revenue Chart */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Weekly Revenue</h2>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={revenueData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="date" />
              <YAxis />
              <Tooltip />
              <Line type="monotone" dataKey="revenue" stroke="#3B82F6" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Orders by Status */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Orders by Status</h2>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={ordersByStatus}
                cx="50%"
                cy="50%"
                innerRadius={60}
                outerRadius={100}
                paddingAngle={5}
                dataKey="value"
                label={({ name, value }) => `${name}: ${value}`}
              >
                {ordersByStatus.map((_, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>

        {/* Top Products */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 lg:col-span-2">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Top Products</h2>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={topProducts} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis type="number" />
              <YAxis dataKey="name" type="category" width={150} />
              <Tooltip />
              <Bar dataKey="sales" fill="#3B82F6" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
}
