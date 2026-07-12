import { Navigate } from "react-router-dom"
import { useAuth } from "@/contexts/AuthContext"

export function AdminGuard({ children }: { children: React.ReactNode }) {
  const { user } = useAuth()

  if (user?.role !== "admin") {
    return <Navigate to="/mail/inbox" replace />
  }

  return <>{children}</>
}
