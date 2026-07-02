import { lazy, Suspense } from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "@/contexts/AuthContext";
import { ThemeProvider } from "@/contexts/ThemeContext";
import { Toaster } from "@/components/ui/sonner";
import Login from "@/pages/Login";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import MainLayout from "@/layout/MainLayout";
import UserLayout from "@/layout/UserLayout";
import "./App.css";

const Dashboard = lazy(() => import("@/pages/Dashboard"));
const Users = lazy(() => import("@/pages/Users"));
const Domains = lazy(() => import("@/pages/Domains"));
const Settings = lazy(() => import("@/pages/mail/Settings"));
const Inbox = lazy(() => import("@/pages/mail/Inbox"));
const Compose = lazy(() => import("@/pages/mail/Compose"));

function PageLoader() {
  return <div className="flex items-center justify-center h-full p-8 text-muted-foreground">Loading...</div>;
}

function App() {
  return (
    <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
      <AuthProvider>
        <Router>
          <Suspense fallback={<PageLoader />}>
            <Routes>
              <Route path="/login" element={<Login />} />

              {/* Admin Routes */}
              <Route element={<ProtectedRoute allowedRoles={['admin']} />}>
                <Route element={<MainLayout />}>
                  <Route path="/" element={<Dashboard />} />
                  <Route path="/users" element={<Users />} />
                  <Route path="/domains" element={<Domains />} />
                </Route>
              </Route>

              {/* Webmail Routes */}
              <Route element={<ProtectedRoute allowedRoles={['user', 'admin']} />}>
                 <Route path="/mail" element={<UserLayout />}>
                    <Route index element={<Navigate to="inbox" replace />} />
                    <Route path="inbox" element={<Inbox />} />
                    <Route path="compose" element={<Compose />} />
                    <Route path="sent" element={<div className="p-4">Sent (Coming Soon)</div>} />
                    <Route path="settings" element={<Settings />} />
                 </Route>
              </Route>

              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
          <Toaster />
        </Router>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
