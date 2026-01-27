import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/button";

export default function Dashboard() {
  const { user, logout } = useAuth();

  return (
    <div className="p-8">
      <h1 className="text-3xl font-bold mb-4">Dashboard</h1>
      <p className="mb-4">Welcome back, {user?.username}!</p>
      <Button onClick={logout} variant="destructive">
        Logout
      </Button>
    </div>
  );
}
