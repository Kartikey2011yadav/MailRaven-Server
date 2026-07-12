import { useEffect, useState } from "react";
import { MessageAPI } from "@/lib/api";
import type { MessageSummary, MessageFull } from "@/lib/api";
import { formatDistanceToNow } from "date-fns";
import { Inbox as InboxIcon, Mail, Reply, Forward } from "lucide-react";
import { useNavigate } from "react-router-dom";
import DOMPurify from "dompurify";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

interface MailFolderProps {
  folder: string;
  emptyTitle?: string;
  emptyDescription?: string;
}

export default function MailFolder({ folder, emptyTitle, emptyDescription }: MailFolderProps) {
  const navigate = useNavigate();
  const [messages, setMessages] = useState<MessageSummary[]>([]);
  const [selectedMessage, setSelectedMessage] = useState<MessageFull | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    setSelectedMessage(null);
    MessageAPI.list({ limit: 50, mailbox: folder })
      .then((res) => {
        const data = res.data;
        setMessages(Array.isArray(data) ? data : data.messages || []);
      })
      .catch(() => setMessages([]))
      .finally(() => setLoading(false));
  }, [folder]);

  async function selectMessage(msg: MessageSummary) {
    try {
      const res = await MessageAPI.get(msg.id);
      setSelectedMessage(res.data);
      if (!msg.read_state) {
        await MessageAPI.update(msg.id, { read_state: true });
        setMessages((prev) =>
          prev.map((m) => (m.id === msg.id ? { ...m, read_state: true } : m))
        );
      }
    } catch {
      // ignore
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  if (messages.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-center">
        <InboxIcon className="h-10 w-10 text-muted-foreground/30 mb-3" />
        <p className="text-sm font-medium text-muted-foreground">{emptyTitle || "No messages"}</p>
        <p className="text-xs text-muted-foreground/70 mt-1">{emptyDescription || `Your ${folder.toLowerCase()} is empty`}</p>
      </div>
    );
  }

  return (
    <div className="flex-1 -m-4">
      <ResizablePanelGroup orientation="horizontal" className="h-[calc(100vh-7rem)]">
        <ResizablePanel defaultSize={35} minSize={25}>
          <ScrollArea className="h-full">
            <div className="divide-y divide-border/30">
              {messages.map((msg) => (
                <button
                  key={msg.id}
                  onClick={() => selectMessage(msg)}
                  className={`w-full text-left px-4 py-3 transition-colors hover:bg-accent/50 ${
                    selectedMessage?.id === msg.id ? "bg-accent/70" : ""
                  } ${!msg.read_state ? "font-medium" : ""}`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm truncate">{msg.sender}</span>
                    <span className="text-[10px] text-muted-foreground shrink-0">
                      {formatDistanceToNow(new Date(msg.received_at), { addSuffix: true })}
                    </span>
                  </div>
                  <p className="text-sm truncate mt-0.5">{msg.subject || "(no subject)"}</p>
                  <p className="text-xs text-muted-foreground truncate mt-0.5">{msg.snippet}</p>
                  {!msg.read_state && (
                    <div className="absolute left-1 top-1/2 -translate-y-1/2 h-2 w-2 rounded-full bg-primary" />
                  )}
                </button>
              ))}
            </div>
          </ScrollArea>
        </ResizablePanel>

        <ResizableHandle withHandle />

        <ResizablePanel defaultSize={65} minSize={40}>
          {selectedMessage ? (
            <ScrollArea className="h-full">
              <div className="p-6">
                <div className="mb-4">
                  <h2 className="text-lg font-semibold">{selectedMessage.subject || "(no subject)"}</h2>
                  <div className="flex items-center gap-2 mt-2">
                    <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center">
                      <Mail className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <p className="text-sm font-medium">{selectedMessage.sender}</p>
                      <p className="text-xs text-muted-foreground">
                        {new Date(selectedMessage.received_at).toLocaleString()}
                      </p>
                    </div>
                    <div className="ml-auto flex items-center gap-1">
                      {selectedMessage.spf_result && (
                        <Badge variant="outline" className="text-[10px]">
                          SPF: {selectedMessage.spf_result}
                        </Badge>
                      )}
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() => navigate(`/mail/compose?replyTo=${selectedMessage.id}`)}
                        title="Reply"
                      >
                        <Reply className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() => navigate(`/mail/compose?forward=${selectedMessage.id}`)}
                        title="Forward"
                      >
                        <Forward className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </div>
                </div>
                <div
                  className="prose prose-sm dark:prose-invert max-w-none"
                  dangerouslySetInnerHTML={{
                    __html: DOMPurify.sanitize(selectedMessage.body || ""),
                  }}
                />
              </div>
            </ScrollArea>
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground text-sm">
              Select a message to read
            </div>
          )}
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}
