import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import type { MessageSummary } from "@/lib/api"
import { formatDistanceToNow } from "date-fns"
import { cn } from "@/lib/utils"

interface MessageListProps {
  items: MessageSummary[]
  selectedId?: string | null
  onSelect: (id: string) => void
}

export function MessageList({ items, selectedId, onSelect }: MessageListProps) {
  return (
    <ScrollArea className="h-screen py-2">
      <div className="flex flex-col gap-2 p-4 pt-0">
        {items.map((item) => (
          <button
            key={item.id}
            className={cn(
              "flex flex-col items-start gap-2 rounded-lg border p-3 text-left text-sm transition-all hover:bg-accent",
              selectedId === item.id && "bg-muted"
            )}
            onClick={() => onSelect(item.id)}
          >
            <div className="flex w-full flex-col gap-1">
              <div className="flex items-center">
                <div className="flex items-center gap-2">
                  <div className="font-semibold">{item.sender}</div>
                  {!item.read_state && (
                    <span className="flex h-2 w-2 rounded-full bg-blue-600" />
                  )}
                </div>
                <div
                  className={cn(
                    "ml-auto text-xs",
                    selectedId === item.id
                      ? "text-foreground"
                      : "text-muted-foreground"
                  )}
                >
                  {formatDistanceToNow(new Date(item.received_at), {
                    addSuffix: true,
                  })}
                </div>
              </div>
              <div className="text-xs font-medium">{item.subject}</div>
            </div>
            <div className="line-clamp-2 text-xs text-muted-foreground">
              {item.snippet.substring(0, 300)}
            </div>
            <div className="flex items-center gap-2">
               {/* Badges for DKIM/SPF could go here */}
               <Badge variant={getVariant(item.spf_result)}>SPF: {item.spf_result}</Badge>
            </div>
          </button>
        ))}
        {items.length === 0 && (
             <div className="text-center text-muted-foreground py-10">No messages found.</div>
        )}
      </div>
    </ScrollArea>
  )
}

function getVariant(status?: string): "default" | "secondary" | "destructive" | "outline" {
  if (status === "pass") return "outline" // Greenish in future
  if (status === "fail") return "destructive"
  return "secondary"
}
