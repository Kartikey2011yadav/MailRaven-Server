import { useEffect, useState } from "react"
import { MessageAPI, type MessageSummary, type MessageFull } from "@/lib/api"
import { MessageList } from "@/components/mail/message-list"
import { MessageDisplay } from "@/components/mail/message-display"
import { toast } from "sonner"
import { Loader2 } from "lucide-react"

import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable"

export default function Inbox() {
  const [messages, setMessages] = useState<MessageSummary[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [selectedMessage, setSelectedMessage] = useState<MessageFull | null>(null)
  const [isLoadingList, setIsLoadingList] = useState(true)
  const [isLoadingDetail, setIsLoadingDetail] = useState(false)

  // Load Message List
  useEffect(() => {
    loadMessages()
  }, [])

  async function loadMessages() {
    try {
      setIsLoadingList(true)
      const { data } = await MessageAPI.list({ limit: 50 })
      setMessages(data.messages || [])
    } catch (error) {
      console.error(error)
      toast.error("Failed to load inbox")
    } finally {
      setIsLoadingList(false)
    }
  }

  // Load Message Detail
  useEffect(() => {
    if (!selectedId) {
        setSelectedMessage(null)
        return
    }
    loadMessageDetail(selectedId)
  }, [selectedId])

  async function loadMessageDetail(id: string) {
      try {
          setIsLoadingDetail(true)
          const { data } = await MessageAPI.get(id)
          setSelectedMessage(data)
          
          if (!data.read_state) {
              // Mark as read silently
              await MessageAPI.update(id, { read_state: true })
              // Update local list state
              setMessages(prev => prev.map(m => m.id === id ? { ...m, read_state: true } : m))
          }
      } catch (error) {
          console.error(error)
          toast.error("Failed to load message")
      } finally {
          setIsLoadingDetail(false)
      }
  }

  return (
    <ResizablePanelGroup orientation="horizontal" className="h-full max-h-[calc(100vh-8rem)] items-stretch rounded-lg border">
      <ResizablePanel defaultSize={30} minSize={20}>
         {isLoadingList ? (
             <div className="flex justify-center p-8"><Loader2 className="animate-spin" /></div>
         ) : (
             <MessageList 
                items={messages} 
                selectedId={selectedId}
                onSelect={setSelectedId}
             />
         )}
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize={70}>
         {isLoadingDetail ? (
             <div className="flex justify-center p-8"><Loader2 className="animate-spin" /></div>
         ) : (
             <MessageDisplay message={selectedMessage} />
         )}
      </ResizablePanel>
    </ResizablePanelGroup>
  )
}
