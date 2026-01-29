import { useEffect, useState } from "react";
import { DomainAPI, Domain } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { CreateDomainDialog } from "@/components/domains/CreateDomainDialog";
import { toast } from "sonner";
import { Loader2, Plus, Trash2, Globe, ShieldCheck } from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

export default function Domains() {
  const [domains, setDomains] = useState<Domain[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchDomains = async () => {
    try {
      setLoading(true);
      const res = await DomainAPI.list();
      const data = res.data;
      if (Array.isArray(data)) {
        setDomains(data);
      } else if ('domains' in data) {
        setDomains(data.domains);
      } else {
        setDomains([]);
      }
    } catch (error) {
      console.error(error);
      toast.error("Failed to load domains");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDomains();
  }, []);

  const handleDelete = async (domain: string) => {
    try {
      await DomainAPI.delete(domain);
      toast.success("Domain deleted");
      fetchDomains();
    } catch {
      toast.error("Failed to delete domain");
    }
  };

  return (
    <div className="p-8 space-y-8">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Domains</h1>
          <p className="text-muted-foreground">Manage hosted email domains and DKIM settings.</p>
        </div>
        <CreateDomainDialog onSuccess={fetchDomains}>
          <Button>
            <Plus className="mr-2 h-4 w-4" /> Add Domain
          </Button>
        </CreateDomainDialog>
      </div>

      <div className="border rounded-md">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Domain Name</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>DKIM Selector</TableHead>
              <TableHead>Created At</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center">
                  <Loader2 className="h-6 w-6 animate-spin mx-auto" />
                </TableCell>
              </TableRow>
            ) : domains.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center">
                  No domains found. Add your first domain to get started.
                </TableCell>
              </TableRow>
            ) : (
              domains.map((domain) => (
                <TableRow key={domain.id}>
                  <TableCell className="font-medium flex items-center gap-2">
                    <Globe className="h-4 w-4 text-muted-foreground" />
                    {domain.name}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2 text-green-600">
                       <ShieldCheck className="h-4 w-4" />
                       <span className="text-sm font-medium">Active</span>
                    </div>
                  </TableCell>
                   <TableCell>
                    {domain.dkim_selector || "default"}
                  </TableCell>
                  <TableCell>
                    {new Date(domain.created_at).toLocaleDateString()}
                  </TableCell>
                  <TableCell className="text-right">
                    <AlertDialog>
                      <AlertDialogTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </AlertDialogTrigger>
                      <AlertDialogContent>
                        <AlertDialogHeader>
                          <AlertDialogTitle>Delete Domain?</AlertDialogTitle>
                          <AlertDialogDescription>
                            Are you sure you want to delete <span className="font-bold">{domain.name}</span>? 
                            This will also delete all users and emails associated with this domain.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                         <AlertDialogFooter>
                          <AlertDialogCancel>Cancel</AlertDialogCancel>
                          <AlertDialogAction
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            onClick={() => handleDelete(domain.name)}
                          >
                            Delete
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
