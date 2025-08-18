'use client';

import { useEffect, useState, useContext } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Calendar, User, MapPin, Globe, Edit, Github, Twitter, Linkedin, Instagram, Facebook, Youtube, MessageCircle, Send } from 'lucide-react';
import BlogCard from '../components/BlogCard';
import { Post } from '../interfaces';
import { axiosInstance } from '../utils/api';
import Loading from '../components/Loading';
import NotFoundPage from '../not-found';
import { imageService } from '../services/imageService';
import { useAlert } from '../components/CustomAlert';
import { useAuth } from '../contexts/authContext';
import { z } from 'zod';

export interface UserProfile {
  username: string;
  first_name?: string;
  last_name?: string;
  avatar?: string;
  bio?: string;
  role: string;
  location?: string;
  website?: string;
  joined_at?: string;
  followers: number;
  following: number;
  social_media?: {
    github?: string;
    twitter?: string;
    linkedin?: string;
    instagram?: string;
    facebook?: string;
    youtube?: string;
    discord?: string;
    telegram?: string;
  };
}

export interface UserProfileResponse {
  success: boolean;
  message: string;
  data: {
    user: UserProfile;
    posts: {
      posts: Post[];
      meta: {
        total: number;
        hasNextPage: boolean;
        page: number;
        limit: number;
        totalPage: number;
      };
    };
  };
}

interface EditProfileForm {
  first_name: string;
  last_name: string;
  bio: string;
  avatar?: string;
  location: string;
  website: string;
  github: string;
  twitter: string;
  linkedin: string;
  instagram: string;
  facebook: string;
  youtube: string;
  discord: string;
  telegram: string;
}

const socialSchema = z.object({
  github: z.string().trim().optional(),
  twitter: z.string().trim().optional(),
  linkedin: z.string().trim().optional(),
  instagram: z.string().trim().optional(),
  facebook: z.string().trim().optional(),
  youtube: z.string().trim().optional(),
  discord: z.string().trim().optional(),
  telegram: z.string().trim().optional(),
  website: z.string().trim().optional(),
});

const trimPrefixes = (value: string) => value.replace(/^https?:\/\//, '').replace(/^www\./, '');

const isValidUrl = (value: string) => {
  try {
    const withProto = value.startsWith('http://') || value.startsWith('https://') ? value : `https://${value}`;
    const u = new URL(withProto);
    return Boolean(u.hostname);
  } catch {
    return false;
  }
};

const validateSocialFrontend = (data: {
  github?: string;
  twitter?: string;
  linkedin?: string;
  instagram?: string;
  facebook?: string;
  youtube?: string;
  discord?: string;
  telegram?: string;
  website?: string;
}): Record<string, string> => {
  const errors: Record<string, string> = {};

  const github = (data.github || '').trim();
  if (github) {
    let v = trimPrefixes(github);
    const idx = v.indexOf('github.com/');
    if (idx !== -1) v = v.split('github.com/')[1].replace(/\/.*/, '');
    const re = /^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$/;
    if (!(re.test(v) && v.length >= 1 && v.length <= 39 && !v.includes('--') && !v.startsWith('-') && !v.endsWith('-')))
      errors.github = 'invalid GitHub username format';
  }

  const twitter = (data.twitter || '').trim();
  if (twitter) {
    let v = trimPrefixes(twitter);
    if (v.includes('twitter.com/')) v = v.split('twitter.com/')[1].replace(/\/.*/, '');
    else if (v.includes('x.com/')) v = v.split('x.com/')[1].replace(/\/.*/, '');
    const re = /^[a-zA-Z0-9_]{1,15}$/;
    if (!re.test(v)) errors.twitter = 'invalid Twitter username format';
  }

  const linkedin = (data.linkedin || '').trim();
  if (linkedin) {
    let v = trimPrefixes(linkedin);
    const re = /^[a-zA-Z0-9_-]{3,100}$/;
    if (v.includes('linkedin.com/in/')) {
      let u = v.split('linkedin.com/in/')[1];
      u = u.split('?')[0].split('#')[0].replace(/\/$/, '');
      if (!re.test(u)) errors.linkedin = 'invalid LinkedIn username format in URL';
    } else {
      if (!re.test(v)) errors.linkedin = 'invalid LinkedIn username format';
    }
  }

  const instagram = (data.instagram || '').trim();
  if (instagram) {
    const re = /^[a-zA-Z0-9._]{1,30}$/;
    if (!re.test(instagram)) errors.instagram = 'invalid Instagram username format';
  }

  const facebook = (data.facebook || '').trim();
  if (facebook) {
    const re = /^[a-zA-Z0-9.]{5,50}$/;
    if (facebook.includes('facebook.com')) {
      if (!isValidUrl(facebook)) errors.facebook = 'invalid Facebook URL';
    } else if (!re.test(facebook)) {
      errors.facebook = 'invalid Facebook username format';
    }
  }

  const youtube = (data.youtube || '').trim();
  if (youtube) {
    const re = /^[a-zA-Z0-9_-]{3,30}$/;
    if (youtube.includes('youtube.com') || youtube.includes('youtu.be')) {
      if (!isValidUrl(youtube)) errors.youtube = 'invalid YouTube URL';
    } else if (!re.test(youtube)) {
      errors.youtube = 'invalid YouTube username format';
    }
  }

  const discord = (data.discord || '').trim();
  if (discord) {
    const re = /^[a-zA-Z0-9_]{2,32}(?:#[0-9]{4})?$/;
    if (!re.test(discord)) errors.discord = 'invalid Discord username format';
  }

  const telegram = (data.telegram || '').trim();
  if (telegram) {
    const re = /^[a-zA-Z][a-zA-Z0-9_]{4,31}$/;
    if (!re.test(telegram)) errors.telegram = 'invalid Telegram username format';
  }

  const website = (data.website || '').trim();
  if (website) {
    const withProto = website.startsWith('http://') || website.startsWith('https://') ? website : `https://${website}`;
    if (!withProto.includes('.') || !isValidUrl(withProto)) errors.website = 'invalid website URL format';
  }

  return errors;
};

export default function UserProfileClient({ initialProfileData }: { initialProfileData: UserProfileResponse }) {
  const { success, error: showError, warning } = useAlert();
  const { user: currentUser } = useAuth();
  const [profileData, setProfileData] = useState<UserProfileResponse | null>(initialProfileData);
  const [isOwnProfile, setIsOwnProfile] = useState<boolean>(false);
  const [supportedRegions, setSupportedRegions] = useState<string[]>([]);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [selectedImageFile, setSelectedImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string>('');
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadedImageData, setUploadedImageData] = useState<{ url: string; id: string } | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const user = profileData!.data.user;
  const posts = profileData!.data.posts;
  const displayName = user.first_name && user.last_name ? `${user.first_name} ${user.last_name}` : user.username;

  // Check if current user can edit this profile
  useEffect(() => {
    console.log("Current User", currentUser);
    console.log("User", user);
    if (currentUser && user) {
      setIsOwnProfile(currentUser.username === user.username);
    }
  }, [currentUser, user]);

  const [editForm, setEditForm] = useState<EditProfileForm>({
    first_name: user.first_name || '',
    last_name: user.last_name || '',
    bio: user.bio || '',
    avatar: user.avatar || '',
    location: user.location || '',
    website: user.website || '',
    github: user.social_media?.github || '',
    twitter: user.social_media?.twitter || '',
    linkedin: user.social_media?.linkedin || '',
    instagram: user.social_media?.instagram || '',
    facebook: user.social_media?.facebook || '',
    youtube: user.social_media?.youtube || '',
    discord: user.social_media?.discord || '',
    telegram: user.social_media?.telegram || '',
  });

  useEffect(() => {
    const fetchRegions = async () => {
      try {
        const response = await axiosInstance.get('/user/regions');
        setSupportedRegions(response.data.data || []);
      } catch (error) {}
    };
    fetchRegions();
  }, []);

  const loadMorePosts = () => {};

  const handleEditProfile = async () => {
    setIsUpdating(true);
    try {
      let avatarUrl = editForm.avatar;

      const validation = socialSchema.safeParse({
        github: editForm.github,
        twitter: editForm.twitter,
        linkedin: editForm.linkedin,
        instagram: editForm.instagram,
        facebook: editForm.facebook,
        youtube: editForm.youtube,
        discord: editForm.discord,
        telegram: editForm.telegram,
        website: editForm.website,
      });
      if (!validation.success) {
        warning('Invalid input, please check your social links.');
        setIsUpdating(false);
        return;
      }
      const mirrorErrors = validateSocialFrontend(validation.data);
      if (Object.keys(mirrorErrors).length > 0) {
        setFieldErrors(mirrorErrors);
        setIsUpdating(false);
        return;
      }

      if (uploadedImageData) {
        avatarUrl = uploadedImageData.url;
      }

      const updateData = {
        username: user.username,
        first_name: editForm.first_name,
        last_name: editForm.last_name,
        bio: editForm.bio,
        avatar: avatarUrl,
        location: editForm.location,
        website: editForm.website,
        github: editForm.github,
        twitter: editForm.twitter,
        linkedin: editForm.linkedin,
        instagram: editForm.instagram,
        facebook: editForm.facebook,
        youtube: editForm.youtube,
        discord: editForm.discord,
        telegram: editForm.telegram,
      };

      await axiosInstance.put('/user/update', updateData);

      setProfileData(prev => {
        if (!prev) return prev;
        return {
          ...prev,
          data: {
            ...prev.data,
            user: {
              ...prev.data.user,
              ...updateData,
              social_media: {
                github: editForm.github,
                twitter: editForm.twitter,
                linkedin: editForm.linkedin,
                instagram: editForm.instagram,
                facebook: editForm.facebook,
                youtube: editForm.youtube,
                discord: editForm.discord,
                telegram: editForm.telegram,
              }
            }
          }
        };
      });

      setSelectedImageFile(null);
      setImagePreview('');
      setUploadedImageData(null);
      setIsEditModalOpen(false);
      success('Profile updated successfully!');
    } catch (error: any) {
      if (error?.response?.status === 400) {
        const msg: string = error.response.data?.error || 'Validation failed. Please check your inputs.';
        const newErrors: Record<string, string> = {};
        msg.split(';').forEach((part: string) => {
          const [k, v] = part.split(':').map((s: string) => s?.trim());
          if (k && v) newErrors[k] = v;
        });
        setFieldErrors(newErrors);
        warning(msg);
      } else {
        showError('Failed to update profile. Please try again.');
      }
    } finally {
      setIsUpdating(false);
    }
  };

  const handleRemoveImage = () => {
    if (selectedImageFile) {
      imageService.cleanupPreviewUrl(imagePreview);
    }
    setSelectedImageFile(null);
    setImagePreview('');
    setUploadedImageData(null);
  };

  if (!profileData) {
    return (
      <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-4 sm:py-6 lg:py-8">
        <Loading label="Loading..." className="h-screen" />
      </div>
    );
  }

  return (
    <div className="container mx-auto px-3 sm:px-4 lg:px-6 py-3 sm:py-4 lg:py-6">
      <Card className="mb-4 sm:mb-6 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 border border-slate-700/50 shadow-lg relative">
        {isOwnProfile && (
          <div className="absolute top-3 right-3 z-10">
            <Dialog open={isEditModalOpen} onOpenChange={setIsEditModalOpen}>
              <DialogTrigger asChild>
                <Button variant="outline" size="sm" className="border-orange-500/30 text-orange-300 hover:bg-orange-500/20 text-xs px-2 py-1 h-7">
                  <Edit className="w-3 h-3 mr-1" />
                  Edit
                </Button>
              </DialogTrigger>
              <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                  <DialogTitle className="text-lg">Edit Profile</DialogTitle>
                </DialogHeader>
                <div className="space-y-6">
                  <div className="text-center space-y-4">
                    <div className="flex flex-col items-center space-y-3">
                      <Avatar className="h-20 w-20 border-4 border-orange-500/50 shadow-lg">
                        <AvatarImage src={imagePreview || editForm.avatar || user.avatar} alt="Profile" />
                        <AvatarFallback className="text-2xl bg-gradient-to-br from-orange-500 to-red-500 text-white">
                          {displayName.charAt(0).toUpperCase()}
                        </AvatarFallback>
                      </Avatar>

                      {isUploading && (
                        <div className="w-full max-w-xs">
                          <div className="flex items-center justify-between text-xs text-slate-400 mb-1">
                            <span>Uploading...</span>
                            <span>{uploadProgress}%</span>
                          </div>
                          <div className="w-full bg-slate-700 rounded-full h-2">
                            <div
                              className="bg-orange-500 h-2 rounded-full transition-all duration-300"
                              style={{ width: `${uploadProgress}%` }}
                            />
                          </div>
                        </div>
                      )}

                      <div className="space-y-2">
                        <Label htmlFor="avatar" className="text-sm font-medium cursor-pointer text-orange-300 hover:text-orange-400 transition-colors">
                          Change Profile Picture
                        </Label>
                        <Input
                          id="avatar"
                          type="file"
                          accept="image/*"
                          onChange={async (e) => {
                            const file = e.target.files?.[0];
                            if (file) {
                              const validation = imageService.validateFile(file);
                              if (!validation.isValid) {
                                warning(validation.error || 'Invalid file format');
                                return;
                              }

                              setSelectedImageFile(file);
                              const previewUrl = imageService.createPreviewUrl(file);
                              setImagePreview(previewUrl);

                              setIsUploading(true);
                              setUploadProgress(0);

                              try {
                                const uploadedUrl = await imageService.uploadProfileImage(file, {
                                  onProgress: (event) => {
                                    setUploadProgress(event.progress);
                                  },
                                  onSuccess: () => {},
                                  onError: (error) => {
                                    warning(`Failed to upload image: ${error.message}`);
                                  }
                                });

                                setUploadedImageData({
                                  url: uploadedUrl,
                                  id: uploadedUrl.split('/').pop() || uploadedUrl
                                });

                              } catch (error) {
                                warning('Failed to upload image. Please try again.');
                                setSelectedImageFile(null);
                                setImagePreview('');
                                setUploadedImageData(null);
                              } finally {
                                setIsUploading(false);
                                setUploadProgress(0);
                              }
                            }
                          }}
                          className="hidden"
                        />
                        <p className="text-xs text-slate-400">Click to upload new profile picture</p>

                        {(selectedImageFile || uploadedImageData) && (
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={handleRemoveImage}
                            className="text-xs px-2 py-1 h-6 border-red-500/30 text-red-400 hover:bg-red-500/20"
                          >
                            Remove Image
                          </Button>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <h3 className="text-sm font-semibold text-slate-300 border-b border-slate-700 pb-2">Basic Information</h3>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="first_name" className="text-sm">First Name</Label>
                        <Input
                          id="first_name"
                          value={editForm.first_name}
                          onChange={(e) => setEditForm(prev => ({ ...prev, first_name: e.target.value }))}
                          placeholder="First Name"
                          className="h-9 text-sm"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="last_name" className="text-sm">Last Name</Label>
                        <Input
                          id="last_name"
                          value={editForm.last_name}
                          onChange={(e) => setEditForm(prev => ({ ...prev, last_name: e.target.value }))}
                          placeholder="Last Name"
                          className="h-9 text-sm"
                        />
                      </div>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <h3 className="text-sm font-semibold text-slate-300 border-b border-slate-700 pb-2">Bio</h3>
                    <div className="space-y-2">
                      <Label htmlFor="bio" className="text-sm">Bio</Label>
                      <Textarea
                        id="bio"
                        value={editForm.bio}
                        onChange={(e) => setEditForm(prev => ({ ...prev, bio: e.target.value }))}
                        placeholder="Tell us about yourself..."
                        rows={3}
                        className="text-sm resize-none"
                      />
                    </div>
                  </div>

                  <div className="space-y-4">
                    <h3 className="text-sm font-semibold text-slate-300 border-b border-slate-700 pb-2">Contact Information</h3>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="location" className="text-sm">Location</Label>
                        <Select value={editForm.location} onValueChange={(value) => setEditForm(prev => ({ ...prev, location: value }))}>
                          <SelectTrigger className="h-9 text-sm">
                            <SelectValue placeholder="Select location" />
                          </SelectTrigger>
                          <SelectContent>
                            {supportedRegions.map((region) => (
                              <SelectItem key={region} value={region}>
                                {region}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="website" className="text-sm">Website</Label>
                        <Input
                          id="website"
                          value={editForm.website}
                          onChange={(e) => { setEditForm(prev => ({ ...prev, website: e.target.value })); setFieldErrors(prev => ({ ...prev, website: '' })); }}
                          placeholder="yourwebsite.com"
                          className={`h-9 text-sm ${fieldErrors.website ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                        />
                        {fieldErrors.website && (<p className="text-xs text-red-400">{fieldErrors.website}</p>)}
                      </div>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <h3 className="text-sm font-semibold text-slate-300 border-b border-slate-700 pb-2">Social Media</h3>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="github" className="flex items-center gap-2 text-sm">
                          <Github className="w-4 h-4" />
                          GitHub
                        </Label>
                        <Input
                          id="github"
                          value={editForm.github}
                          onChange={(e) => { setEditForm(prev => ({ ...prev, github: e.target.value })); setFieldErrors(prev => ({ ...prev, github: '' })); }}
                          placeholder="username or URL"
                          className={`h-9 text-sm ${fieldErrors.github ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                        />
                        {fieldErrors.github && (<p className="text-xs text-red-400">{fieldErrors.github}</p>)}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="twitter" className="flex items-center gap-2 text-sm">
                          <Twitter className="w-4 h-4" />
                          Twitter/X
                        </Label>
                        <Input
                          id="twitter"
                          value={editForm.twitter}
                          onChange={(e) => { setEditForm(prev => ({ ...prev, twitter: e.target.value })); setFieldErrors(prev => ({ ...prev, twitter: '' })); }}
                          placeholder="username or URL"
                          className={`h-9 text-sm ${fieldErrors.twitter ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                        />
                        {fieldErrors.twitter && (<p className="text-xs text-red-400">{fieldErrors.twitter}</p>)}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="linkedin" className="flex items-center gap-2 text-sm">
                          <Linkedin className="w-4 h-4" />
                          LinkedIn
                        </Label>
                        <Input
                          id="linkedin"
                          value={editForm.linkedin}
                          onChange={(e) => { setEditForm(prev => ({ ...prev, linkedin: e.target.value })); setFieldErrors(prev => ({ ...prev, linkedin: '' })); }}
                          placeholder="username or URL"
                          className={`h-9 text-sm ${fieldErrors.linkedin ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                        />
                        {fieldErrors.linkedin && (<p className="text-xs text-red-400">{fieldErrors.linkedin}</p>)}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="instagram" className="flex items-center gap-2 text-sm">
                          <Instagram className="w-4 h-4" />
                          Instagram
                        </Label>
                        <Input
                          id="instagram"
                          value={editForm.instagram}
                          onChange={(e) => { setEditForm(prev => ({ ...prev, instagram: e.target.value })); setFieldErrors(prev => ({ ...prev, instagram: '' })); }}
                          placeholder="username or URL"
                          className={`h-9 text-sm ${fieldErrors.instagram ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                        />
                        {fieldErrors.instagram && (<p className="text-xs text-red-400">{fieldErrors.instagram}</p>)}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="facebook" className="flex items-center gap-2 text-sm">
                          <Facebook className="w-4 h-4" />
                          Facebook
                        </Label>
                        <Input
                          id="facebook"
                          value={editForm.facebook}
                          onChange={(e) => { setEditForm(prev => ({ ...prev, facebook: e.target.value })); setFieldErrors(prev => ({ ...prev, facebook: '' })); }}
                          placeholder="username or URL"
                          className={`h-9 text-sm ${fieldErrors.facebook ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                        />
                        {fieldErrors.facebook && (<p className="text-xs text-red-400">{fieldErrors.facebook}</p>)}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="youtube" className="flex items-center gap-2 text-sm">
                          <Youtube className="w-4 h-4" />
                          YouTube
                        </Label>
                        <Input
                          id="youtube"
                          value={editForm.youtube}
                          onChange={(e) => { setEditForm(prev => ({ ...prev, youtube: e.target.value })); setFieldErrors(prev => ({ ...prev, youtube: '' })); }}
                          placeholder="channel name or URL"
                          className={`h-9 text-sm ${fieldErrors.youtube ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                        />
                        {fieldErrors.youtube && (<p className="text-xs text-red-400">{fieldErrors.youtube}</p>)}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="discord" className="flex items-center gap-2 text-sm">
                          <MessageCircle className="w-4 h-4" />
                          Discord
                        </Label>
                        <Input
                          id="discord"
                          value={editForm.discord}
                          onChange={(e) => { setEditForm(prev => ({ ...prev, discord: e.target.value })); setFieldErrors(prev => ({ ...prev, discord: '' })); }}
                          placeholder="username#1234"
                          className={`h-9 text-sm ${fieldErrors.discord ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                        />
                        {fieldErrors.discord && (<p className="text-xs text-red-400">{fieldErrors.discord}</p>)}
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="telegram" className="flex items-center gap-2 text-sm">
                          <Send className="w-4 h-4" />
                          Telegram
                        </Label>
                        <Input
                          id="telegram"
                          value={editForm.telegram}
                          onChange={(e) => { setEditForm(prev => ({ ...prev, telegram: e.target.value })); setFieldErrors(prev => ({ ...prev, telegram: '' })); }}
                          placeholder="username"
                          className={`h-9 text-sm ${fieldErrors.telegram ? 'border-red-500 focus-visible:ring-red-500' : ''}`}
                        />
                        {fieldErrors.telegram && (<p className="text-xs text-red-400">{fieldErrors.telegram}</p>)}
                      </div>
                    </div>
                  </div>

                  <div className="flex justify-end gap-3 pt-6 border-t border-slate-700">
                    <Button
                      variant="outline"
                      onClick={() => setIsEditModalOpen(false)}
                      disabled={isUpdating}
                      size="sm"
                      className="h-9 px-4 text-sm"
                    >
                      Cancel
                    </Button>
                    <Button
                      onClick={handleEditProfile}
                      disabled={isUpdating}
                      size="sm"
                      className="bg-orange-600 hover:bg-orange-700 h-9 px-6 text-sm"
                    >
                      {isUpdating ? 'Saving...' : 'Save Changes'}
                    </Button>
                  </div>
                </div>
              </DialogContent>
            </Dialog>
          </div>
        )}

        <CardHeader className="text-center pb-2 sm:pb-3 px-3 sm:px-4">
          <div className="flex flex-col items-center space-y-2 sm:space-y-3">
            <Avatar className="h-12 w-12 sm:h-16 sm:w-16 md:h-20 md:w-20 lg:h-24 lg:w-24 border-3 border-orange-500/50 shadow-lg">
              <AvatarImage src={user.avatar} alt={displayName} />
              <AvatarFallback className="text-lg sm:text-xl lg:text-2xl bg-gradient-to-br from-orange-500 to-red-500 text-white">
                {displayName.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>

            <div className="space-y-0.5">
              <CardTitle className="text-lg sm:text-xl lg:text-2xl font-bold text-white leading-tight">{displayName}</CardTitle>
              <p className="text-sm sm:text-base text-orange-300">@{user.username}</p>
              {user.role === 'WRITER_USER' && (
                <Badge variant="secondary" className="mt-1 px-1.5 py-0.5 text-xs bg-orange-500/20 text-orange-300 border-orange-500/30">
                  <User className="w-3 h-3 mr-1" />
                  นักเขียน
                </Badge>
              )}
            </div>

            {user.bio && (
              <p className="text-slate-300 max-w-lg text-center text-xs sm:text-sm leading-relaxed px-2">
                {user.bio}
              </p>
            )}

            <div className="flex flex-col sm:flex-row items-center justify-center gap-1.5 sm:gap-2 md:gap-3 text-slate-400 text-xs">
              {user.location && (
                <div className="flex items-center gap-1">
                  <MapPin className="w-3 h-3 sm:w-3 sm:h-3 text-orange-400" />
                  <span>{user.location}</span>
                </div>
              )}
              {user.website && (
                <div className="flex items-center gap-1">
                  <Globe className="w-3 h-3 sm:w-3 sm:h-3 text-orange-400" />
                  <a
                    href={user.website.startsWith('http') ? user.website : `https://${user.website}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-slate-400 hover:text-orange-400 transition-colors cursor-pointer decoration-dotted"
                  >
                    {user.website.replace(/^https?:\/\//, '').replace(/^www\./, '').split('/')[0]}
                  </a>
                </div>
              )}
              {user.joined_at && (
                <div className="flex items-center gap-1">
                  <Calendar className="w-3 h-3 sm:w-3 sm:h-3 text-orange-400" />
                  <span>Joined {user.joined_at}</span>
                </div>
              )}
            </div>

            {user.social_media && Object.values(user.social_media).some(val => val) && (
              <div className="flex flex-wrap items-center justify-center gap-1.5">
                {user.social_media.github && (
                  <a
                    href={user.social_media.github.startsWith('http') ? user.social_media.github : `https://github.com/${user.social_media.github.replace(/^https?:\/\//, '').replace(/^github\.com\//, '')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-slate-400 hover:text-orange-400 transition-colors"
                  >
                    <Github className="w-3.5 h-3.5" />
                  </a>
                )}
                {user.social_media.twitter && (
                  <a
                    href={user.social_media.twitter.startsWith('http') ? user.social_media.twitter : `https://twitter.com/${user.social_media.twitter.replace(/^https?:\/\//, '').replace(/^twitter\.com\//, '').replace(/^x\.com\//, '')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-slate-400 hover:text-orange-400 transition-colors"
                  >
                    <Twitter className="w-3.5 h-3.5" />
                  </a>
                )}
                {user.social_media.linkedin && (
                  <a
                    href={user.social_media.linkedin.startsWith('http') ? user.social_media.linkedin : `https://linkedin.com/in/${user.social_media.linkedin.replace(/^https?:\/\//, '').replace(/^linkedin\.com\/in\//, '')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-slate-400 hover:text-orange-400 transition-colors"
                  >
                    <Linkedin className="w-3.5 h-3.5" />
                  </a>
                )}
                {user.social_media.instagram && (
                  <a
                    href={user.social_media.instagram.startsWith('http') ? user.social_media.instagram : `https://instagram.com/${user.social_media.instagram.replace(/^https?:\/\//, '').replace(/^instagram\.com\//, '')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-slate-400 hover:text-orange-400 transition-colors"
                  >
                    <Instagram className="w-3.5 h-3.5" />
                  </a>
                )}
                {user.social_media.facebook && (
                  <a
                    href={user.social_media.facebook.startsWith('http') ? user.social_media.facebook : `https://facebook.com/${user.social_media.facebook.replace(/^https?:\/\//, '').replace(/^facebook\.com\//, '')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-slate-400 hover:text-orange-400 transition-colors"
                  >
                    <Facebook className="w-3.5 h-3.5" />
                  </a>
                )}
                {user.social_media.youtube && (
                  <a
                    href={user.social_media.youtube.startsWith('http') ? user.social_media.youtube : `https://youtube.com/@${user.social_media.youtube.replace(/^https?:\/\//, '').replace(/^youtube\.com\/@?/, '')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-slate-400 hover:text-orange-400 transition-colors"
                  >
                    <Youtube className="w-3.5 h-3.5" />
                  </a>
                )}
                {user.social_media.discord && (
                  <span className="text-slate-400 flex items-center gap-1">
                    <MessageCircle className="w-3.5 h-3.5" />
                    <span className="text-xs">{user.social_media.discord}</span>
                  </span>
                )}
                {user.social_media.telegram && (
                  <a
                    href={user.social_media.telegram.startsWith('http') ? user.social_media.telegram : `https://t.me/${user.social_media.telegram.replace(/^https?:\/\//, '').replace(/^t\.me\//, '')}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-slate-400 hover:text-orange-400 transition-colors"
                  >
                    <Send className="w-3.5 h-3.5" />
                  </a>
                )}
              </div>
            )}

          </div>
        </CardHeader>
      </Card>

      <div className="mb-3 sm:mb-4">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1.5 sm:gap-2 mb-3 sm:mb-4">
          <h2 className="text-lg sm:text-xl font-bold text-white">Latest Posts</h2>
          <span className="text-xs sm:text-sm text-slate-400">
            {posts.meta.total} articles
          </span>
        </div>

        {posts && posts.posts && posts.posts.length === 0 ? (
          <Card className="bg-slate-900 border-slate-700/50">
            <CardContent className="text-center py-6 sm:py-8">
              <div className="text-slate-500 mb-1.5 sm:mb-2">
                <Calendar className="w-6 h-6 sm:w-8 sm:h-8 mx-auto" />
              </div>
              <p className="text-xs sm:text-sm text-slate-400">
                ยังไม่มีบทความที่เผยแพร่
              </p>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2 sm:gap-3 lg:gap-4">
            {posts?.posts?.map((post) => (
              <BlogCard key={post.slug} post={post} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
