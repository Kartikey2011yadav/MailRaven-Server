import { useEffect, useState } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import api from "@/lib/api";
import { Users, Mail, Activity, AlertCircle } from "lucide-react";

interface SystemStats {
  users: {
    total: number;
    active: number;
    admin: number;
  };
  emails: {
    total: number;
  };
  queue: {
    pending: number;
    processing: number;
    failed: number;
    completed: number;
  };
}

export default function Dashboard() {
  const { user, logout } = useAuth();
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await api.get('/admin/stats');
        setStats(response.data);
        setError(null);
      } catch (err: any) {
        console.error("Failed to fetch stats:", err);
        if (err.response && err.response.status === 403) {
           setError("You do not have permission to view stats.");
        } else {
           setError("Failed to load system statistics. Ensure the backend is running.");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  return (
    <div className="p-8 space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground">Welcome back, {user?.username}!</p>
        </div>
        <Button onClick={logout} variant="secondary">
          Logout
        </Button>
      </div>

      {error && (
        <div className="p-4 rounded-md bg-destructive/15 text-destructive flex items-center gap-2">
          <AlertCircle className="h-4 w-4" />
          <span>{error}</span>
        </div>
      )}

      {loading ? (
        <div className="text-muted-foreground">Loading stats...</div>
      ) : stats ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Users</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.users.total}</div>
              <p className="text-xs text-muted-foreground">
                {stats.users.active} active, {stats.users.admin} admins
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Emails</CardTitle>
              <Mail className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.emails.total}</div>
            </CardContent>
          </Card>
           <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Queue Pending</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.queue.pending}</div>
               <p className="text-xs text-muted-foreground">
                {stats.queue.processing} processing
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Queue Failed</CardTitle>
              <AlertCircle className="h-4 w-4 text-destructive" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-destructive">{stats.queue.failed}</div>
               <p className="text-xs text-muted-foreground">
                {stats.queue.completed} sent successfully
              </p>
            </CardContent>
          </Card>
        </div>
      ) : null}
    </div>
  );
}
