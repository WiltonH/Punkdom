"use client"

import {
  CircleCheckIcon,
  InfoIcon,
  Loader2Icon,
  OctagonXIcon,
  TriangleAlertIcon,
} from "lucide-react"
import { useTheme } from "next-themes"
import { Toaster as Sonner, type ToasterProps } from "sonner"

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "dark" } = useTheme()
  const sonnerTheme = theme === "paper" ? "light" : theme

  return (
    <Sonner
      {...props}
      theme={sonnerTheme as ToasterProps["theme"]}
      className="toaster group"
      richColors={false}
      icons={{
        success: <CircleCheckIcon className="size-4" />,
        info: <InfoIcon className="size-4" />,
        warning: <TriangleAlertIcon className="size-4" />,
        error: <OctagonXIcon className="size-4" />,
        loading: <Loader2Icon className="size-4 animate-spin" />,
      }}
      toastOptions={{
        classNames: {
          toast: "group toast group-[.toaster]:rounded-[var(--punkdom-radius)] group-[.toaster]:border group-[.toaster]:border-[var(--punkdom-border)] group-[.toaster]:bg-[var(--punkdom-surface)] group-[.toaster]:text-[var(--punkdom-text)] group-[.toaster]:shadow-[var(--punkdom-shadow)] group-[.toaster]:backdrop-blur-none",
          title: "group-[.toast]:text-xs group-[.toast]:font-medium group-[.toast]:text-[var(--punkdom-text)]",
          description: "group-[.toast]:text-xs group-[.toast]:text-[var(--punkdom-text-muted)]",
          actionButton: "group-[.toast]:rounded-[var(--punkdom-radius)] group-[.toast]:border group-[.toast]:border-[var(--punkdom-border)] group-[.toast]:bg-[var(--punkdom-active)] group-[.toast]:text-[var(--punkdom-text)]",
          cancelButton: "group-[.toast]:rounded-[var(--punkdom-radius)] group-[.toast]:border group-[.toast]:border-[var(--punkdom-border)] group-[.toast]:bg-[var(--punkdom-surface-2)] group-[.toast]:text-[var(--punkdom-text-muted)]",
          closeButton: "group-[.toast]:border-[var(--punkdom-border)] group-[.toast]:bg-[var(--punkdom-surface-2)] group-[.toast]:text-[var(--punkdom-text-muted)]",
          error: "group toast group-[.toaster]:border-[var(--punkdom-danger-border)] group-[.toaster]:bg-[var(--punkdom-surface)] group-[.toaster]:text-[var(--punkdom-text)] group-[.toaster]:[--normal-border:var(--punkdom-danger-border)] group-[.toaster]:[--normal-bg:var(--punkdom-surface)] group-[.toaster]:[--normal-text:var(--punkdom-text)]",
          success: "group toast group-[.toaster]:border-[var(--punkdom-border)] group-[.toaster]:bg-[var(--punkdom-surface)] group-[.toaster]:text-[var(--punkdom-text)]",
          info: "group toast group-[.toaster]:border-[var(--punkdom-border)] group-[.toaster]:bg-[var(--punkdom-surface)] group-[.toaster]:text-[var(--punkdom-text)]",
          warning: "group toast group-[.toaster]:border-[var(--punkdom-warning-bg)] group-[.toaster]:bg-[var(--punkdom-surface)] group-[.toaster]:text-[var(--punkdom-text)]",
        },
      }}
      style={
        {
          "--normal-bg": "var(--punkdom-surface)",
          "--normal-text": "var(--punkdom-text)",
          "--normal-border": "var(--punkdom-border)",
          "--border-radius": "var(--punkdom-radius)",
        } as React.CSSProperties
      }
    />
  )
}

export { Toaster }
