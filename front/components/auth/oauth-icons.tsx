import type { OAuthProvider } from "@/components/auth/oauth-buttons";

type IconProps = { className?: string };

// Official brand marks. Google and Microsoft keep their brand colors; GitHub is
// monochrome and inherits the current text color via `fill="currentColor"`.

function GoogleIcon({ className }: IconProps) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <path
        fill="#4285F4"
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.27-4.74 3.27-8.1Z"
      />
      <path
        fill="#34A853"
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84A11 11 0 0 0 12 23Z"
      />
      <path
        fill="#FBBC05"
        d="M5.84 14.1a6.6 6.6 0 0 1 0-4.2V7.06H2.18a11 11 0 0 0 0 9.88l3.66-2.84Z"
      />
      <path
        fill="#EA4335"
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1A11 11 0 0 0 2.18 7.06l3.66 2.84C6.71 7.3 9.14 5.38 12 5.38Z"
      />
    </svg>
  );
}

function GitHubIcon({ className }: IconProps) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
    >
      <path d="M12 1A11 11 0 0 0 8.52 22.44c.55.1.75-.24.75-.53v-1.86c-3.06.67-3.71-1.47-3.71-1.47-.5-1.27-1.22-1.61-1.22-1.61-1-.68.08-.67.08-.67 1.1.08 1.69 1.14 1.69 1.14.98 1.69 2.58 1.2 3.21.92.1-.71.38-1.2.69-1.48-2.44-.28-5.01-1.22-5.01-5.44 0-1.2.43-2.18 1.13-2.95-.11-.28-.49-1.4.11-2.91 0 0 .92-.3 3.02 1.13a10.5 10.5 0 0 1 5.5 0c2.1-1.43 3.02-1.13 3.02-1.13.6 1.51.22 2.63.11 2.91.7.77 1.13 1.75 1.13 2.95 0 4.23-2.58 5.16-5.03 5.43.4.34.74 1 .74 2.02v3c0 .3.2.64.76.53A11 11 0 0 0 12 1Z" />
    </svg>
  );
}

function MicrosoftIcon({ className }: IconProps) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#F25022" d="M2 2h9.5v9.5H2Z" />
      <path fill="#7FBA00" d="M12.5 2H22v9.5h-9.5Z" />
      <path fill="#00A4EF" d="M2 12.5h9.5V22H2Z" />
      <path fill="#FFB900" d="M12.5 12.5H22V22h-9.5Z" />
    </svg>
  );
}

const ICONS: Record<OAuthProvider, (props: IconProps) => React.ReactElement> = {
  google: GoogleIcon,
  github: GitHubIcon,
  microsoft: MicrosoftIcon,
};

/** OAuthIcon renders the brand mark for the given provider. */
export default function OAuthIcon({
  provider,
  className,
}: {
  provider: OAuthProvider;
  className?: string;
}) {
  const Icon = ICONS[provider];
  return <Icon className={className} />;
}
