import type { ReactNode } from "react";

export function Panel({ title, label, children, className = "" }: {
  title: string; label: string; children: ReactNode; className?: string;
}) {
  return <section className={`panel ${className}`} aria-label={title}>
    <div className="panel-heading"><span className="eyebrow">{label}</span><h2>{title}</h2></div>
    {children}
  </section>;
}
