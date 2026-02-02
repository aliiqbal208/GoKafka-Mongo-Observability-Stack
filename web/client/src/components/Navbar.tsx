import { Link } from 'react-router-dom';
import { ShoppingCart, User, LogOut, Package } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useCart } from '../context/CartContext';

export default function Navbar() {
  const { user, logout } = useAuth();
  const { totalItems } = useCart();

  return (
    <nav className="bg-white shadow-md sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <Link to="/" className="flex items-center space-x-2">
              <Package className="h-8 w-8 text-indigo-600" />
              <span className="text-xl font-bold text-gray-900">ShopHub</span>
            </Link>
          </div>

          <div className="flex items-center space-x-4">
            <Link
              to="/products"
              className="text-gray-600 hover:text-indigo-600 px-3 py-2 rounded-md text-sm font-medium"
            >
              Products
            </Link>

            <Link to="/cart" className="relative p-2 text-gray-600 hover:text-indigo-600">
              <ShoppingCart className="h-6 w-6" />
              {totalItems > 0 && (
                <span className="absolute -top-1 -right-1 bg-indigo-600 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center">
                  {totalItems}
                </span>
              )}
            </Link>

            {user ? (
              <div className="flex items-center space-x-4">
                <Link
                  to="/orders"
                  className="text-gray-600 hover:text-indigo-600 px-3 py-2 rounded-md text-sm font-medium"
                >
                  My Orders
                </Link>
                <span className="text-gray-700 text-sm">{user.name}</span>
                <button
                  onClick={logout}
                  className="p-2 text-gray-600 hover:text-red-600"
                  title="Logout"
                >
                  <LogOut className="h-5 w-5" />
                </button>
              </div>
            ) : (
              <Link
                to="/login"
                className="flex items-center space-x-1 bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700"
              >
                <User className="h-4 w-4" />
                <span>Login</span>
              </Link>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}
