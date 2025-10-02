import UserProfileClient, { UserProfileResponse } from './user-profile-client';
import { axiosInstanceServer } from '../../lib/api-server';
import { notFound } from 'next/navigation';

export default async function UserProfilePage({ params }: { params: Promise<{ username: string }> }) {
  const { username: rawUsername } = await params;
  const username = decodeURIComponent(rawUsername).replace(/^@/, '');

  if (!username) {
    notFound();
  }

  try {
    const res = await axiosInstanceServer.get(`/user/profile/${username}/posts`);
    if (res.status !== 200) {
      notFound();
    }
    const initialProfileData: UserProfileResponse = res.data;
    return <UserProfileClient initialProfileData={initialProfileData} />;
  } catch (e: any) {
    if (e?.response?.status === 404) {
      notFound();
    }
    notFound();
  }
}
