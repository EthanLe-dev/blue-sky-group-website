import "server-only";
import { cache } from "react";
import { read } from "./catalog";
import type { CompanyProfile } from "@/components/HomeFooter";

// Each request uses one settings fetch, even when many server-rendered cards need its contact links.
export const readCompanySettings = cache(() =>
  read<{ company_profile?: CompanyProfile; response_sla?: { message: string } }>("/site-settings").catch(() => null),
);
