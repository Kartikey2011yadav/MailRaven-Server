import * as React from "react"
import {
  BookOpen,
  Frame,
  GalleryVerticalEnd,
  Map,
  PieChart,
  SquareTerminal,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { NavSystem } from "@/components/nav-system"
import { NavUser } from "@/components/nav-user"
import { TeamSwitcher } from "@/components/team-switcher"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"

// This is sample data.
const data = {
  user: {
    name: "admin",
    email: "admin@mailraven.local",
    avatar: "/avatars/shadcn.jpg",
  },
  teams: [
    {
      name: "MailRaven",
      logo: GalleryVerticalEnd,
      plan: "Enterprise",
    },
  ],
  navMain: [
    {
      title: "System",
      url: "#",
      icon: SquareTerminal,
      items: [
        {
          title: "Dashboard",
          url: "/",
        },
        {
          title: "Domains",
          url: "/domains",
        },
        {
          title: "Users",
          url: "/users",
        },
      ],
    },
    {
      title: "Users",
      url: "/users",
      icon: BookOpen, // Or better icon
      items: [
        {
          title: "Manage Users",
          url: "/users",
        },
      ],
    },
    {
      title: "Domains",
      url: "/domains",
      icon: Frame,
      items: [
         {
          title: "All Domains",
          url: "/domains",
         }
      ]
    },
  ],
  navSystem: [

    {
      name: "System Logs",
      url: "/logs",
      icon: Frame,
    },
    {
      name: "Queue Manager",
      url: "/queue",
      icon: PieChart,
    },
    {
      name: "DKIM/DMARC",
      url: "/security",
      icon: Map,
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <TeamSwitcher teams={data.teams} />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSystem projects={data.navSystem} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
