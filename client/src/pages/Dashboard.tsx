import { useEffect, useState } from "react";
import { useAuth } from "@/contexts/AuthContext";
import { GlassCard, GlassCardContent, GlassCardHeader, GlassCardTitle } from "@/components/ui/glass-card";
import api from "@/lib/api";
import { Users, Mail, Activity, AlertCircle, TrendingUp } from "lucide-react";
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from "recharts";

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
  const { user } = useAuth();
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await api.get('/admin/stats');
        setStats(response.data);
        setError(null);
      } catch (err: unknown) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        if ((err as any).response?.status === 403) {
          setError("You do not have permission to view stats.");
        } else {
          setError("Failed to load statistics.");
        }
      } finally {
        setLoading(false);
      }
    };
    fetchStats();
  }, []);

  // Mock chart data (will be replaced with real API in future)
  const chartData = [
    { day: "Mon", emails: 12 },
    { day: "Tue", emails: 19 },
    { day: "Wed", emails: 8 },
    { day: "Thu", emails: 24 },
    { day: "Fri", emails: 15 },
    { day: "Sat", emails: 6 },
    { day: "Sun", emails: 3 },
  ];

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <GlassCard key={i} className="h-[120px] animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-sm text-muted-foreground">Welcome back, {user?.username}</p>
      </div>

      {error && (
        <div className="p-3 rounded-lg bg-destructive/10 text-destructive text-sm flex items-center gap-2">
          <AlertCircle className="h-4 w-4" />
          <span>{error}</span>
        </div>
      )}

      {stats && (
        <>
          {/* Stat Cards */}
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <StatCard
              title="Total Users"
              value={stats.users.total}
              subtitle={`${stats.users.admin} admins`}
              icon={Users}
            />
            <StatCard
              title="Total Emails"
              value={stats.emails.total}
              icon={Mail}
            />
            <StatCard
              title="Queue Pending"
              value={stats.queue.pending}
              subtitle={`${stats.queue.processing} processing`}
              icon={Activity}
            />
            <StatCard
              title="Queue Failed"
              value={stats.queue.failed}
              subtitle={`${stats.queue.completed} delivered`}
              icon={AlertCircle}
              variant={stats.queue.failed > 0 ? "destructive" : "default"}
            />
          </div>

          {/* Chart */}
          <GlassCard>
            <GlassCardHeader>
              <div className="flex items-center gap-2">
                <TrendingUp className="h-4 w-4 text-primary" />
                <GlassCardTitle>Email Volume (7 days)</GlassCardTitle>
              </div>
            </GlassCardHeader>
            <GlassCardContent>
              <div className="h-[200px]">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData}>
                    <defs>
                      <linearGradient id="emailGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="hsl(234, 89%, 63%)" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="hsl(234, 89%, 63%)" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <XAxis
                      dataKey="day"
                      stroke="hsl(220, 9%, 43%)"
                      fontSize={12}
                      tickLine={false}
                      axisLine={false}
                    />
                    <YAxis
                      stroke="hsl(220, 9%, 43%)"
                      fontSize={12}
                      tickLine={false}
                      axisLine={false}
                    />
                    <Tooltip
                      contentStyle={{
                        background: "hsl(225, 22%, 9%)",
                        border: "1px solid rgba(255,255,255,0.08)",
                        borderRadius: "8px",
                        fontSize: "12px",
                      }}
                    />
                    <Area
                      type="monotone"
                      dataKey="emails"
                      stroke="hsl(234, 89%, 63%)"
                      strokeWidth={2}
                      fill="url(#emailGradient)"
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </GlassCardContent>
          </GlassCard>
        </>
      )}
    </div>
  );
}

function StatCard({
  title,
  value,
  subtitle,
  icon: Icon,
  variant = "default",
}: {
  title: string;
  value: number;
  subtitle?: string;
  icon: React.ComponentType<{ className?: string }>;
  variant?: "default" | "destructive";
}) {
  return (
    <GlassCard>
      <GlassCardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <GlassCardTitle className="text-sm font-medium text-muted-foreground">{title}</GlassCardTitle>
        <Icon className={`h-4 w-4 ${variant === "destructive" ? "text-destructive" : "text-primary/70"}`} />
      </GlassCardHeader>
      <GlassCardContent>
        <div className={`text-2xl font-bold ${variant === "destructive" ? "text-destructive" : ""}`}>
          {value.toLocaleString()}
        </div>
        {subtitle && (
          <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>
        )}
      </GlassCardContent>
    </GlassCard>
  );
}
