import { Users as UsersIcon } from 'lucide-react';

export default function Users() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Users</h1>
        <p className="text-gray-500">Manage customer accounts</p>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-12 text-center">
        <UsersIcon className="w-16 h-16 text-gray-300 mx-auto mb-4" />
        <h2 className="text-xl font-semibold text-gray-600 mb-2">User Management</h2>
        <p className="text-gray-400">
          User management features coming soon. You'll be able to view, edit, and manage customer accounts here.
        </p>
      </div>
    </div>
  );
}
