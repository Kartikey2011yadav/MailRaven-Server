import { useState, useEffect } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import * as z from "zod"
import { Loader2, Send } from "lucide-react"
import { toast } from "sonner"

import { MessageAPI } from "@/lib/api"
import { Button } from "@/components/ui/button"
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { GlassCard, GlassCardContent, GlassCardHeader, GlassCardTitle } from "@/components/ui/glass-card"

const formSchema = z.object({
  to: z.string().email("Invalid email address"),
  subject: z.string().min(1, "Subject is required"),
  body: z.string().min(1, "Message body is required"),
})

export default function Compose() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [isSending, setIsSending] = useState(false)
  const [loadingContext, setLoadingContext] = useState(false)

  const replyTo = searchParams.get("replyTo")
  const forward = searchParams.get("forward")

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      to: "",
      subject: "",
      body: "",
    },
  })

  useEffect(() => {
    const messageId = replyTo || forward
    if (!messageId) return

    setLoadingContext(true)
    MessageAPI.get(messageId)
      .then((res) => {
        const msg = res.data
        if (replyTo) {
          form.setValue("to", msg.sender)
          form.setValue("subject", msg.subject.startsWith("Re:") ? msg.subject : `Re: ${msg.subject}`)
          form.setValue("body", `\n\n---\nOn ${new Date(msg.received_at).toLocaleString()}, ${msg.sender} wrote:\n> ${(msg.body || "").replace(/<[^>]*>/g, "").split("\n").join("\n> ")}`)
        } else if (forward) {
          form.setValue("subject", msg.subject.startsWith("Fwd:") ? msg.subject : `Fwd: ${msg.subject}`)
          form.setValue("body", `\n\n---\nForwarded message from ${msg.sender} on ${new Date(msg.received_at).toLocaleString()}:\n\n${(msg.body || "").replace(/<[^>]*>/g, "")}`)
        }
      })
      .catch(() => {
        toast.error("Failed to load original message")
      })
      .finally(() => setLoadingContext(false))
  }, [replyTo, forward, form])

  async function onSubmit(values: z.infer<typeof formSchema>) {
    setIsSending(true)
    try {
      await MessageAPI.send(values)
      toast.success("Message sent")
      navigate("/mail/inbox")
    } catch {
      toast.error("Failed to send message")
    } finally {
      setIsSending(false)
    }
  }

  const title = replyTo ? "Reply" : forward ? "Forward" : "Compose"

  return (
    <div className="max-w-2xl mx-auto">
      <GlassCard>
        <GlassCardHeader>
          <GlassCardTitle>{title}</GlassCardTitle>
        </GlassCardHeader>
        <GlassCardContent>
          {loadingContext ? (
            <div className="flex items-center justify-center py-8">
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
            </div>
          ) : (
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                  control={form.control}
                  name="to"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs text-muted-foreground">To</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="recipient@example.com"
                          className="bg-secondary/50 border-border/50"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="subject"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs text-muted-foreground">Subject</FormLabel>
                      <FormControl>
                        <Input
                          placeholder="Enter subject"
                          className="bg-secondary/50 border-border/50"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="body"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel className="text-xs text-muted-foreground">Message</FormLabel>
                      <FormControl>
                        <Textarea
                          placeholder="Type your message..."
                          className="min-h-[250px] bg-secondary/50 border-border/50 font-mono text-sm"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <div className="flex justify-end gap-2 pt-2">
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => navigate(-1)}
                  >
                    Cancel
                  </Button>
                  <Button type="submit" size="sm" disabled={isSending}>
                    {isSending ? (
                      <Loader2 className="h-4 w-4 animate-spin mr-1" />
                    ) : (
                      <Send className="h-4 w-4 mr-1" />
                    )}
                    {isSending ? "Sending..." : "Send"}
                  </Button>
                </div>
              </form>
            </Form>
          )}
        </GlassCardContent>
      </GlassCard>
    </div>
  )
}
