import { useEffect, useState } from "react";
import { UserAPI } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { CreateUserDialog } from "@/components/users/CreateUserDialog";
import { DeleteUserDialog } from "@/components/users/DeleteUserDialog";
import { toast } from "sonner";
import { Loader2, Plus, Shield, User } from "lucide-react";

interface User {
  email: string;
  role: string;
  created_at?: string;
  last_login_at?: string;
}

export default function Users() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      const res = await UserAPI.list();
      // The API returns { users: [...], total: ... } or just array depending on backend implementation. 
      // Based on my backend check:
      // internal/adapters/http/handlers/admin_users.go:ListUsers likely returns a list or struct.
      // I should double check the response format.
      // Assuming it returns { users: [] } or just [] based on typical patterns.
      // Let's assume response.data is the payload.
      // If the backend returns `json.NewEncoder(w).Encode(users)` it's an array.
      // If it returns a wrapped object, I need to adjust.
      const data = res.data;
      setUsers(data.users);
    } catch (error) {
      console.error(error);
      toast.error("Failed to load users");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  return (
    <div className="p-8 space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Users</h1>
          <p className="text-muted-foreground">Manage system access and roles.</p>
        </div>
        <CreateUserDialog onSuccess={fetchUsers}>
          <Button>
            <Plus className="mr-2 h-4 w-4" /> Add User
          </Button>
        </CreateUserDialog>
      </div>

      <div className="border rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Last Login</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={4} className="h-24 text-center">
                  <Loader2 className="h-6 w-6 animate-spin mx-auto" />
                </TableCell>
              </TableRow>
            ) : users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="h-24 text-center">
                  No users found.
                </TableCell>
              </TableRow>
            ) : (
              users.map((user) => (
                <TableRow key={user.email}>
                  <TableCell className="font-medium">{user.email}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                       {user.role === 'ADMIN' ? <Shield className="h-3 w-3" /> : <User className="h-3 w-3" />}
                       {user.role}
                    </div>
                  </TableCell>
                  <TableCell>
                    {user.last_login_at ? new Date(user.last_login_at).toLocaleString() : "Never"}
                  </TableCell>
                  <TableCell className="text-right">
                    <DeleteUserDialog 
                      email={user.email} 
                      onSuccess={fetchUsers} 
                      disabled={user.role === "ADMIN"} // Prevent deleting admins for safety (optional logic)
                    />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
