import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";

interface ProtectedRouteProps {
  allowedRoles?: string[];
}

export const ProtectedRoute = ({ allowedRoles }: ProtectedRouteProps) => {
  const { isAuthenticated, isLoading, user } = useAuth();

  if (isLoading) {
    return <div>Loading...</div>; // Or a proper spinner
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (allowedRoles && user?.role && !allowedRoles.includes(user.role)) {
     // Simplistic Redirect Strategy:
     // If user is trying to access Admin routes (allowedRoles=['admin']) but is 'user', send to /mail
     if (user.role === 'user') {
       return <Navigate to="/mail" replace />;
     }
     return <Navigate to="/" replace />;
  }

  return <Outlet />;
};
