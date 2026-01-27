import * as React from "react"
import {
  AudioWaveform,
  BookOpen,
  Bot,
  Command,
  Frame,
  GalleryVerticalEnd,
  Map,
  PieChart,
  Settings2,
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
      title: "Dashboard",
      url: "/",
      icon: SquareTerminal,
      isActive: true,
      items: [
        {
          title: "Overview",
          url: "/",
        },
        {
          title: "Real-time Stats",
          url: "/stats",
        },
      ],
    },
    {
      title: "Domains",
      url: "/domains",
      icon: Bot,
      items: [
        {
          title: "List Domains",
          url: "/domains",
        },
        {
          title: "Add Domain",
          url: "/domains/new",
        },
      ],
    },
    {
      title: "Accounts",
      url: "/accounts",
      icon: BookOpen,
      items: [
        {
          title: "Mailboxes",
          url: "/accounts",
        },
        {
          title: "Aliases",
          url: "/aliases",
        },
      ],
    },
    {
      title: "Settings",
      url: "/settings",
      icon: Settings2,
      items: [
        {
          title: "General",
          url: "/settings",
        },
        {
          title: "Security",
          url: "/settings/security",
        },
      ],
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
