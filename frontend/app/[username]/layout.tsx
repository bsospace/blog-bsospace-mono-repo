import { Metadata } from "next";
import { ReactNode } from "react";
import { axiosInstanceServer } from "../../lib/api-server";
import envConfig from '../configs/envConfig';

export const dynamic = 'force-dynamic';
export const revalidate = 0;

const SITE_NAME = 'Blog Space Blog';
const SITE_URL = envConfig.domain;

export default function UserProfileLayout({ children }: { children: ReactNode }) {
  return (
    <div>
      {children}
    </div>
  );
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ username: string }>;
}): Promise<Metadata> {
  const { username: rawParam } = await params;
  const username = decodeURIComponent(rawParam || '').replace(/^@/, '');

  // Defaults
  let displayName = username || 'User';
  let bio: string | undefined;
  let avatar: string | undefined;
  let twitterHandle: string | undefined;

  try {
    const res = await axiosInstanceServer.get(`/user/profile/${username}/posts`);
    if (res.status === 200) {
      const payload = res.data?.data;
      const user = payload?.user ?? payload;
      const firstName = user?.first_name || user?.firstName;
      const lastName = user?.last_name || user?.lastName;
      displayName = firstName && lastName ? `${firstName} ${lastName}` : (user?.username || displayName);
      bio = user?.bio || undefined;
      avatar = user?.avatar || undefined;
      twitterHandle = user?.social_media?.twitter || undefined;
    }
  } catch {
    // keep defaults
  }

  const title = `${SITE_NAME} - ${displayName} (@${username || 'user'})`;
  const description = bio || `Read posts and profile of ${displayName} (@${username}) on ${SITE_NAME}.`;
  const canonical = `${SITE_URL}/${rawParam || username}`;

  return {
    title,
    description,
    alternates: { canonical },
    robots: { index: true, follow: true },
    keywords: [displayName, username, 'blog', 'articles', 'profile', 'technology', 'developer'],
    openGraph: {
      siteName: SITE_NAME,
      title,
      description,
      url: canonical,
      type: 'profile',
      images: avatar ? [{ url: avatar }] : undefined,
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: avatar ? [avatar] : undefined,
      creator: twitterHandle ? (twitterHandle.startsWith('@') ? twitterHandle : `@${twitterHandle}`) : undefined,
    },
    metadataBase: new URL(SITE_URL),
    authors: [{ name: displayName }],
  };
}