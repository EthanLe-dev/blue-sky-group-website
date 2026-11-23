export const companyContact = {
  hotline: "0989 876 847",
  phoneHref: "tel:0989876847",
  zaloHref: "https://zalo.me/0989876847",
  facebookHref: "https://www.facebook.com/profile.php?id=61561766096180",
};


export type ContactLinks = typeof companyContact;
export function resolveContact(company?: { hotline?: string | null; zalo_url?: string | null; is_demo?: boolean } | null): ContactLinks {
  const phone = company?.hotline?.replace(/[^+0-9]/g, "") || "";
  return {
    ...companyContact,
    ...(company?.is_demo === false && /^\+?\d{7,15}$/.test(phone)
      ? { hotline: company.hotline!, phoneHref: `tel:${phone}` } : {}),
    ...(company?.is_demo === false && company.zalo_url && /^https:\/\/zalo\.me\//i.test(company.zalo_url)
      ? { zaloHref: company.zalo_url } : {}),
  };
}
