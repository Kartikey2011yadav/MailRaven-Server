import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import type { MessageFull } from "@/lib/api"
import { format } from "date-fns"
import { Reply, ReplyAll, Forward } from "lucide-react"
import DOMPurify from "dompurify"

interface MessageDisplayProps {
  message: MessageFull | null
}

export function MessageDisplay({ message }: MessageDisplayProps) {
  if (!message) {
    return (
      <div className="p-8 text-center text-muted-foreground">
        No message selected
      </div>
    )
  }

  const cleanHTML = DOMPurify.sanitize(message.body)

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-start p-4">
        <div className="flex items-start gap-4 text-sm">
          <Avatar>
            <AvatarImage alt={message.sender} />
            <AvatarFallback>
              {message.sender
                .split(" ")
                .map((chunk) => chunk[0])
                .join("")}
            </AvatarFallback>
          </Avatar>
          <div className="grid gap-1">
            <div className="font-semibold">{message.sender}</div>
            <div className="line-clamp-1 text-xs">{message.subject}</div>
            <div className="line-clamp-1 text-xs">
              <span className="font-medium">Reply-To:</span> {message.sender}
            </div>
          </div>
        </div>
        <div className="ml-auto text-xs text-muted-foreground">
          {format(new Date(message.received_at), "PPpp")}
        </div>
      </div>
      <Separator />
      
      <div className="p-4 flex gap-2">
         <Button size="sm" variant="outline"><Reply className="w-4 h-4 mr-2"/> Reply</Button>
         <Button size="sm" variant="outline"><ReplyAll className="w-4 h-4 mr-2"/> Reply All</Button>
         <Button size="sm" variant="outline"><Forward className="w-4 h-4 mr-2"/> Forward</Button>
      </div>
      <Separator />

      <div className="flex-1 overflow-y-auto p-4 whitespace-pre-wrap">
          <div dangerouslySetInnerHTML={{ __html: cleanHTML }} />
      </div>
    </div>
  )
}
