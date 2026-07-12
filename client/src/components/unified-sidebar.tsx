import {
  Inbox,
  Send,
  Trash2,
  Settings,
  Archive,
  FileEdit,
  LayoutDashboard,
  Globe,
  Users,
  Activity,
  Server,
  PenSquare,
  Mail,
  LogOut,
} from "lucide-react"

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarGroupContent,
  SidebarSeparator,
} from "@/components/ui/sidebar"
import { useAuth } from "@/contexts/AuthContext"
import { Link, useLocation, useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { ModeToggle } from "@/components/ModeToggle"

const mailNavItems = [
  { title: "Inbox", url: "/mail/inbox", icon: Inbox },
  { title: "Drafts", url: "/mail/drafts", icon: FileEdit },
  { title: "Sent", url: "/mail/sent", icon: Send },
  { title: "Archive", url: "/mail/archive", icon: Archive },
  { title: "Trash", url: "/mail/trash", icon: Trash2 },
]

const adminNavItems = [
  { title: "Dashboard", url: "/admin/dashboard", icon: LayoutDashboard },
  { title: "Domains", url: "/admin/domains", icon: Globe },
  { title: "Users", url: "/admin/users", icon: Users },
  { title: "Queue", url: "/admin/queue", icon: Activity },
  { title: "System", url: "/admin/system", icon: Server },
]

export function UnifiedSidebar() {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const isAdmin = user?.role === "admin"

  const initials = user?.username
    ? user.username.charAt(0).toUpperCase()
    : "U"

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <div className="flex items-center gap-2 px-2 py-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
            <Mail className="h-4 w-4 text-primary-foreground" />
          </div>
          <div className="flex flex-col group-data-[collapsible=icon]:hidden">
            <span className="text-sm font-bold gradient-text">MailRaven</span>
            <span className="text-[10px] text-muted-foreground">Self-hosted email</span>
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent>
        {/* Compose Button */}
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  className="bg-primary text-primary-foreground hover:bg-primary/90 font-medium"
                >
                  <Link to="/mail/compose">
                    <PenSquare />
                    <span>Compose</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        {/* Mail Section */}
        <SidebarGroup>
          <SidebarGroupLabel>Mail</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {mailNavItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    isActive={location.pathname === item.url}
                  >
                    <Link to={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        {/* Admin Section — only visible to admins */}
        {isAdmin && (
          <>
            <SidebarSeparator />
            <SidebarGroup>
              <SidebarGroupLabel>Admin</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  {adminNavItems.map((item) => (
                    <SidebarMenuItem key={item.title}>
                      <SidebarMenuButton
                        asChild
                        isActive={location.pathname === item.url}
                      >
                        <Link to={item.url}>
                          <item.icon />
                          <span>{item.title}</span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </>
        )}
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              isActive={location.pathname === "/mail/settings"}
            >
              <Link to="/mail/settings">
                <Settings />
                <span>Settings</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <SidebarSeparator />
        <div className="flex items-center gap-2 px-2 py-2">
          <Avatar className="h-7 w-7">
            <AvatarFallback className="text-xs bg-primary/10 text-primary">
              {initials}
            </AvatarFallback>
          </Avatar>
          <div className="flex flex-1 flex-col group-data-[collapsible=icon]:hidden">
            <span className="text-xs font-medium truncate">{user?.username}</span>
            <span className="text-[10px] text-muted-foreground capitalize">{user?.role}</span>
          </div>
          <div className="flex items-center gap-1 group-data-[collapsible=icon]:hidden">
            <ModeToggle />
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => { logout(); navigate("/login"); }}
              title="Logout"
            >
              <LogOut className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
