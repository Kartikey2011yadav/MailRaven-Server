import * as React from "react"
import {
  Inbox,
  Send,
  Trash2,
  Settings,
  Archive,
  FileEdit,
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
} from "@/components/ui/sidebar"
import { NavUser } from "@/components/nav-user"
import { useAuth } from "@/contexts/AuthContext"
import { Link, useLocation } from "react-router-dom"
import { TeamSwitcher } from "@/components/team-switcher"
import { GalleryVerticalEnd } from "lucide-react"

// Mock data for user sidebar
const data = {
    teams: [
    {
      name: "MailRaven",
      logo: GalleryVerticalEnd,
      plan: "Webmail",
    },
  ],
}

export function MailSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const location = useLocation();
  const { user } = useAuth();
  
  const navItems = [
    { title: "Inbox", url: "/mail/inbox", icon: Inbox },
    { title: "Drafts", url: "/mail/drafts", icon: FileEdit },
    { title: "Sent", url: "/mail/sent", icon: Send },
    { title: "Archive", url: "/mail/archive", icon: Archive },
    { title: "Trash", url: "/mail/trash", icon: Trash2 },
  ];

  const settingsItems = [
    { title: "Settings", url: "/mail/settings", icon: Settings },
  ];

  const userData = {
      name: user?.username || "User",
      email: user?.username || "user@example.com",
      avatar: "",
  }

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <TeamSwitcher teams={data.teams} />
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
           <SidebarGroupLabel>Mail</SidebarGroupLabel>
           <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild isActive={location.pathname.startsWith(item.url)}>
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

        <SidebarGroup>
          <SidebarGroupLabel>Account</SidebarGroupLabel>
          <SidebarGroupContent>
               <SidebarMenu>
              {settingsItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild isActive={location.pathname === item.url}>
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
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={userData} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
