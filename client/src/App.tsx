import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "@/contexts/AuthContext";
import { ThemeProvider } from "@/contexts/ThemeContext";
import { Toaster } from "@/components/ui/sonner";
import Login from "@/pages/Login";
import Dashboard from "@/pages/Dashboard";
import Users from "@/pages/Users";
import Domains from "@/pages/Domains"; // Import Domains page
import { ProtectedRoute } from "@/components/ProtectedRoute";
import MainLayout from "@/layout/MainLayout";
import UserLayout from "@/layout/UserLayout";
import Settings from "@/pages/mail/Settings";
import Inbox from "@/pages/mail/Inbox";
import "./App.css";

function App() {
  return (
    <ThemeProvider defaultTheme="system" storageKey="vite-ui-theme">
      <AuthProvider>
        <Router>
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
                  <Route path="sent" element={<div className="p-4">Sent (Coming Soon)</div>} />
                  <Route path="settings" element={<Settings />} />
               </Route>
            </Route>

            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
          <Toaster />
        </Router>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App;
