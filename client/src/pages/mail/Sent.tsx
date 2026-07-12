import MailFolder from "./MailFolder";

export default function Sent() {
  return <MailFolder folder="Sent" emptyTitle="No sent messages" emptyDescription="Messages you send will appear here" />;
}
