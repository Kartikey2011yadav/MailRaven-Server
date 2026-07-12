import { lazy, Suspense } from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "@/contexts/AuthContext";
import { ThemeProvider } from "@/contexts/ThemeContext";
import { Toaster } from "@/components/ui/sonner";
import Login from "@/pages/Login";
import SetupWizard from "@/pages/setup/SetupWizard";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { AdminGuard } from "@/components/AdminGuard";
import UnifiedLayout from "@/layout/UnifiedLayout";

const Dashboard = lazy(() => import("@/pages/Dashboard"));
const Users = lazy(() => import("@/pages/Users"));
const Domains = lazy(() => import("@/pages/Domains"));
const Settings = lazy(() => import("@/pages/mail/Settings"));
const Inbox = lazy(() => import("@/pages/mail/Inbox"));
const Compose = lazy(() => import("@/pages/mail/Compose"));
const Sent = lazy(() => import("@/pages/mail/Sent"));
const Drafts = lazy(() => import("@/pages/mail/Drafts"));
const Archive = lazy(() => import("@/pages/mail/Archive"));
const Trash = lazy(() => import("@/pages/mail/Trash"));

function PageLoader() {
  return (
    <div className="flex items-center justify-center h-full p-8">
      <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
    </div>
  );
}

function App() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="mailraven-theme">
      <AuthProvider>
        <Router>
          <Suspense fallback={<PageLoader />}>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route path="/setup" element={<SetupWizard />} />

              {/* All authenticated routes — single unified layout */}
              <Route element={<ProtectedRoute />}>
                <Route element={<UnifiedLayout />}>
                  {/* Mail routes (all authenticated users) */}
                  <Route path="/mail/inbox" element={<Inbox />} />
                  <Route path="/mail/compose" element={<Compose />} />
                  <Route path="/mail/sent" element={<Sent />} />
                  <Route path="/mail/drafts" element={<Drafts />} />
                  <Route path="/mail/archive" element={<Archive />} />
                  <Route path="/mail/trash" element={<Trash />} />
                  <Route path="/mail/settings" element={<Settings />} />

                  {/* Admin routes (admin role required) */}
                  <Route path="/admin/dashboard" element={<AdminGuard><Dashboard /></AdminGuard>} />
                  <Route path="/admin/users" element={<AdminGuard><Users /></AdminGuard>} />
                  <Route path="/admin/domains" element={<AdminGuard><Domains /></AdminGuard>} />
                  <Route path="/admin/queue" element={<AdminGuard><PlaceholderPage title="Queue Monitor" /></AdminGuard>} />
                  <Route path="/admin/system" element={<AdminGuard><PlaceholderPage title="System" /></AdminGuard>} />
                </Route>
              </Route>

              {/* Default redirect */}
              <Route path="/" element={<Navigate to="/mail/inbox" replace />} />
              <Route path="/mail" element={<Navigate to="/mail/inbox" replace />} />
              <Route path="*" element={<Navigate to="/mail/inbox" replace />} />
            </Routes>
          </Suspense>
          <Toaster richColors position="bottom-right" />
        </Router>
      </AuthProvider>
    </ThemeProvider>
  );
}

function PlaceholderPage({ title }: { title: string }) {
  return (
    <div className="flex flex-1 items-center justify-center">
      <div className="text-center">
        <h2 className="text-lg font-medium text-muted-foreground">{title}</h2>
        <p className="text-sm text-muted-foreground/70 mt-1">Coming soon</p>
      </div>
    </div>
  );
}

export default App;
