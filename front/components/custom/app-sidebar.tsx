import { QuoteIcon, ReceiptEuroIcon, WrenchIcon } from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "../ui/sidebar";

const items = [
  {
    title: "Devis",
    url: "/quote",
    icon: QuoteIcon,
  },
  {
    title: "Factures",
    url: "/invoice",
    icon: ReceiptEuroIcon,
  },
  {
    title: "Clients",
    url: "/clients",
    icon: QuoteIcon,
  },
  {
    title: "Test",
    url: "/test",
    icon: WrenchIcon,
  },
];

export default function AppSidebar() {
  return (
    <>
      <Sidebar>
        <SidebarContent className="bg-primary-foreground text-primary">
          <SidebarGroup>
            <SidebarGroupLabel>Application</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {items.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton asChild>
                      <a href={item.url}>
                        <item.icon />
                        <span>{item.title}</span>
                      </a>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
      </Sidebar>
    </>
  );
}
