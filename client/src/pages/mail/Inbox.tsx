import MailFolder from "./MailFolder";

export default function Inbox() {
  return <MailFolder folder="INBOX" emptyTitle="Inbox is empty" emptyDescription="New messages will appear here" />;
}
