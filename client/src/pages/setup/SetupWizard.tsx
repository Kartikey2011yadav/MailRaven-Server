import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Mail, ArrowRight, ArrowLeft, Check, Copy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { SetupAPI } from "@/lib/api";
import type { DNSRecord } from "@/lib/api";

const STEPS = ["Welcome", "Domain", "Admin", "DNS", "Done"];

export default function SetupWizard() {
  const navigate = useNavigate();
  const [step, setStep] = useState(0);
  const [isLoading, setIsLoading] = useState(false);

  // Form data
  const [domain, setDomain] = useState("");
  const [smtpHostname, setSmtpHostname] = useState("");
  const [adminEmail, setAdminEmail] = useState("");
  const [adminPassword, setAdminPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [dnsRecords, setDnsRecords] = useState<DNSRecord[]>([]);

  const next = () => setStep((s) => Math.min(s + 1, STEPS.length - 1));
  const prev = () => setStep((s) => Math.max(s - 1, 0));

  async function handleComplete() {
    if (adminPassword !== confirmPassword) {
      toast.error("Passwords do not match");
      return;
    }
    if (adminPassword.length < 8) {
      toast.error("Password must be at least 8 characters");
      return;
    }

    setIsLoading(true);
    try {
      const res = await SetupAPI.complete({
        domain,
        admin_email: adminEmail,
        admin_password: adminPassword,
        smtp_hostname: smtpHostname || `mail.${domain}`,
      });
      setDnsRecords(res.data.dns_records);
      next();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Setup failed";
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    toast.success("Copied to clipboard");
  }

  return (
    <div className="relative flex items-center justify-center min-h-screen overflow-hidden bg-background px-4">
      <div className="absolute inset-0 bg-gradient-to-br from-primary/5 via-background to-purple-500/5" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,rgba(99,102,241,0.08),transparent_50%)]" />

      <div className="relative w-full max-w-lg">
        {/* Progress */}
        <div className="flex justify-center gap-2 mb-6">
          {STEPS.map((_, i) => (
            <div
              key={i}
              className={`h-1.5 rounded-full transition-all duration-300 ${
                i <= step
                  ? "w-8 bg-primary"
                  : "w-4 bg-muted"
              }`}
            />
          ))}
        </div>

        <div className="glass-card rounded-2xl p-8 min-h-[400px] flex flex-col">
          <AnimatePresence mode="wait">
            <motion.div
              key={step}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
              className="flex-1 flex flex-col"
            >
              {step === 0 && (
                <WelcomeStep />
              )}
              {step === 1 && (
                <DomainStep
                  domain={domain}
                  setDomain={setDomain}
                  smtpHostname={smtpHostname}
                  setSmtpHostname={setSmtpHostname}
                />
              )}
              {step === 2 && (
                <AdminStep
                  domain={domain}
                  adminEmail={adminEmail}
                  setAdminEmail={setAdminEmail}
                  adminPassword={adminPassword}
                  setAdminPassword={setAdminPassword}
                  confirmPassword={confirmPassword}
                  setConfirmPassword={setConfirmPassword}
                />
              )}
              {step === 3 && (
                <DNSStep records={dnsRecords} onCopy={copyToClipboard} />
              )}
              {step === 4 && (
                <DoneStep />
              )}
            </motion.div>
          </AnimatePresence>

          {/* Navigation */}
          <div className="flex justify-between mt-6 pt-4 border-t border-border/30">
            {step > 0 && step < 3 ? (
              <Button variant="ghost" size="sm" onClick={prev}>
                <ArrowLeft className="h-4 w-4 mr-1" /> Back
              </Button>
            ) : (
              <div />
            )}

            {step === 0 && (
              <Button size="sm" onClick={next}>
                Get Started <ArrowRight className="h-4 w-4 ml-1" />
              </Button>
            )}
            {step === 1 && (
              <Button size="sm" onClick={next} disabled={!domain}>
                Next <ArrowRight className="h-4 w-4 ml-1" />
              </Button>
            )}
            {step === 2 && (
              <Button
                size="sm"
                onClick={handleComplete}
                disabled={!adminEmail || !adminPassword || isLoading}
              >
                {isLoading ? "Setting up..." : "Complete Setup"}
                {!isLoading && <Check className="h-4 w-4 ml-1" />}
              </Button>
            )}
            {step === 3 && (
              <Button size="sm" onClick={next}>
                Continue <ArrowRight className="h-4 w-4 ml-1" />
              </Button>
            )}
            {step === 4 && (
              <Button size="sm" onClick={() => navigate("/login")}>
                Go to Login <ArrowRight className="h-4 w-4 ml-1" />
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function WelcomeStep() {
  return (
    <div className="flex flex-col items-center justify-center flex-1 text-center">
      <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-primary/10 glow-sm mb-5">
        <Mail className="h-7 w-7 text-primary" />
      </div>
      <h2 className="text-2xl font-bold gradient-text mb-2">Welcome to MailRaven</h2>
      <p className="text-muted-foreground text-sm max-w-sm">
        Let's set up your self-hosted email server. This wizard will configure your domain,
        create an admin account, and show you the DNS records to set up.
      </p>
    </div>
  );
}

function DomainStep({
  domain, setDomain, smtpHostname, setSmtpHostname,
}: {
  domain: string;
  setDomain: (v: string) => void;
  smtpHostname: string;
  setSmtpHostname: (v: string) => void;
}) {
  return (
    <div className="flex flex-col flex-1">
      <h2 className="text-lg font-semibold mb-1">Configure Domain</h2>
      <p className="text-sm text-muted-foreground mb-6">
        Enter your email domain. This is the part after @ in email addresses.
      </p>
      <div className="space-y-4">
        <div>
          <label className="text-xs font-medium text-muted-foreground">Domain Name</label>
          <Input
            placeholder="example.com"
            value={domain}
            onChange={(e) => setDomain(e.target.value)}
            className="mt-1 bg-secondary/50 border-border/50"
          />
          <p className="text-xs text-muted-foreground mt-1">Users will have addresses like user@{domain || "example.com"}</p>
        </div>
        <div>
          <label className="text-xs font-medium text-muted-foreground">SMTP Hostname (optional)</label>
          <Input
            placeholder={domain ? `mail.${domain}` : "mail.example.com"}
            value={smtpHostname}
            onChange={(e) => setSmtpHostname(e.target.value)}
            className="mt-1 bg-secondary/50 border-border/50"
          />
          <p className="text-xs text-muted-foreground mt-1">Defaults to mail.{domain || "example.com"}</p>
        </div>
      </div>
    </div>
  );
}

function AdminStep({
  domain, adminEmail, setAdminEmail, adminPassword, setAdminPassword,
  confirmPassword, setConfirmPassword,
}: {
  domain: string;
  adminEmail: string;
  setAdminEmail: (v: string) => void;
  adminPassword: string;
  setAdminPassword: (v: string) => void;
  confirmPassword: string;
  setConfirmPassword: (v: string) => void;
}) {
  return (
    <div className="flex flex-col flex-1">
      <h2 className="text-lg font-semibold mb-1">Create Admin Account</h2>
      <p className="text-sm text-muted-foreground mb-6">
        This will be the administrator account for managing your server.
      </p>
      <div className="space-y-4">
        <div>
          <label className="text-xs font-medium text-muted-foreground">Admin Email</label>
          <Input
            placeholder={`admin@${domain}`}
            value={adminEmail}
            onChange={(e) => setAdminEmail(e.target.value)}
            className="mt-1 bg-secondary/50 border-border/50"
          />
        </div>
        <div>
          <label className="text-xs font-medium text-muted-foreground">Password</label>
          <Input
            type="password"
            placeholder="Minimum 8 characters"
            value={adminPassword}
            onChange={(e) => setAdminPassword(e.target.value)}
            className="mt-1 bg-secondary/50 border-border/50"
          />
        </div>
        <div>
          <label className="text-xs font-medium text-muted-foreground">Confirm Password</label>
          <Input
            type="password"
            placeholder="Repeat password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            className="mt-1 bg-secondary/50 border-border/50"
          />
          {confirmPassword && confirmPassword !== adminPassword && (
            <p className="text-xs text-destructive mt-1">Passwords do not match</p>
          )}
        </div>
      </div>
    </div>
  );
}

function DNSStep({ records, onCopy }: { records: DNSRecord[]; onCopy: (v: string) => void }) {
  return (
    <div className="flex flex-col flex-1">
      <h2 className="text-lg font-semibold mb-1">Configure DNS Records</h2>
      <p className="text-sm text-muted-foreground mb-4">
        Add these records to your domain's DNS settings for email to work properly.
      </p>
      <div className="space-y-2 overflow-y-auto max-h-[260px]">
        {records.map((record, i) => (
          <div key={i} className="rounded-lg bg-secondary/30 p-3 border border-border/30">
            <div className="flex items-center justify-between mb-1">
              <span className="text-xs font-mono font-bold text-primary">{record.type}</span>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6"
                onClick={() => onCopy(record.value)}
              >
                <Copy className="h-3 w-3" />
              </Button>
            </div>
            <p className="text-xs text-muted-foreground">Name: <span className="font-mono">{record.name}</span></p>
            <p className="text-xs text-muted-foreground break-all">Value: <span className="font-mono">{record.value}</span></p>
          </div>
        ))}
      </div>
    </div>
  );
}

function DoneStep() {
  return (
    <div className="flex flex-col items-center justify-center flex-1 text-center">
      <div className="flex h-14 w-14 items-center justify-center rounded-full bg-green-500/10 mb-5">
        <Check className="h-7 w-7 text-green-500" />
      </div>
      <h2 className="text-2xl font-bold mb-2">Setup Complete!</h2>
      <p className="text-muted-foreground text-sm max-w-sm">
        Your MailRaven server is configured. Don't forget to add the DNS records shown in the
        previous step. You can now log in with your admin credentials.
      </p>
    </div>
  );
}
