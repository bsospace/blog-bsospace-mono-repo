'use client';

import { useEffect, useState, useContext } from 'react';
import { useParams, notFound } from 'next/navigation';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Calendar, Eye, Heart, Clock, User, MapPin, Globe, Edit, Github, Twitter, Linkedin, Instagram, Facebook, Youtube, MessageCircle, Send } from 'lucide-react';
import Link from 'next/link';
import { formatDistanceToNow } from 'date-fns';
import { th } from 'date-fns/locale';
import { axiosInstance } from '../utils/api';
import BlogCard from '../components/BlogCard';
import { Post } from '../interfaces';
import Loading from '../components/Loading';
import { useAuth } from '../contexts/authContext';
import NotFoundPage from '../not-found';
import { imageService } from '../services/imageService';
import { useAlert } from '../components/CustomAlert';

interface UserProfile {
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
  can_edit?: boolean; // Added for can_edit
}

interface UserProfileResponse {
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
  avatar?: string; // Added for file upload preview
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

export default function UserProfilePage() {
  const params = useParams();
  const usernameWithAt = params?.username as string;
  const { user: currentUser, isLoggedIn } = useAuth();
  const { success, error: showError, warning } = useAlert();
  const [profileData, setProfileData] = useState<UserProfileResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [isOwnProfile, setIsOwnProfile] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [editForm, setEditForm] = useState<EditProfileForm>({
    first_name: '',
    last_name: '',
    bio: '',
    avatar: '',
    location: '',
    website: '',
    github: '',
    twitter: '',
    linkedin: '',
    instagram: '',
    facebook: '',
    youtube: '',
    discord: '',
    telegram: '',
  });
  const [supportedRegions, setSupportedRegions] = useState<string[]>([]);
  const [selectedImageFile, setSelectedImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string>('');
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadedImageData, setUploadedImageData] = useState<{ url: string; id: string } | null>(null);

  // Decode URL parameter and split username from @ symbol
  const decodedUsername = decodeURIComponent(usernameWithAt || '');
  const username = decodedUsername?.startsWith('@') ? decodedUsername.slice(1) : decodedUsername;

  // Validate username format
  const isValidUsername = username && username.length > 0 && /^[a-zA-Z0-9._-]+$/.test(username);

  // Fetch supported regions
  useEffect(() => {
    const fetchRegions = async () => {
      try {
        const response = await axiosInstance.get('/user/regions');
        setSupportedRegions(response.data.data || []);
      } catch (error) {
        console.error('Failed to fetch regions:', error);
      }
    };
    fetchRegions();
  }, []);

  // Fetch user profile and posts
  useEffect(() => {
    const fetchUserData = async () => {
      try {
        setLoading(true);
        const response = await axiosInstance.get(`/user/profile/${username}/posts`);
        setProfileData(response.data);
        
        // Use can_edit from backend (now available through optional auth middleware)
        setIsOwnProfile(response.data.data.user.can_edit || false);
        
        // Initialize edit form with current user data
        if (response.data.data.user) {
          const userData = response.data.data.user;
          setEditForm({
            first_name: userData.first_name || '',
            last_name: userData.last_name || '',
            bio: userData.bio || '',
            avatar: userData.avatar || '',
            location: userData.location || '',
            website: userData.website || '',
            github: userData.social_media?.github || '',
            twitter: userData.social_media?.twitter || '',
            linkedin: userData.social_media?.linkedin || '',
            instagram: userData.social_media?.instagram || '',
            facebook: userData.social_media?.facebook || '',
            youtube: userData.social_media?.youtube || '',
            discord: userData.social_media?.discord || '',
            telegram: userData.social_media?.telegram || '',
          });
        }
      } catch (err: any) {
        if (err.response?.status === 404) {
          notFound();
        } else {
          setError('เกิดข้อผิดพลาดในการโหลดข้อมูล');
        }
      } finally {
        setLoading(false);
      }
    };

    if (username) {
      fetchUserData();
    }
  }, [username]);

  const loadMorePosts = () => {
    setCurrentPage(prev => prev + 1);
  };

  const handleEditProfile = async () => {
    setIsUpdating(true);
    try {
      let avatarUrl = editForm.avatar;

      // Use pre-uploaded image data if available
      if (uploadedImageData) {
        avatarUrl = uploadedImageData.url;
      }

      // Update user profile with new avatar URL
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

      // Update local state
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

      // Reset form and close modal
      setSelectedImageFile(null);
      setImagePreview('');
      setUploadedImageData(null);
      setIsEditModalOpen(false);
      
      // Show success message
      success('Profile updated successfully!');
    } catch (error) {
      showError('Failed to update profile. Please try again.');
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

  if (loading) {
    return (
      <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-4 sm:py-6 lg:py-8">
        <Loading label="กำลังโหลดโปรไฟล์..." className="h-screen" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 sm:px-6 lg:px-8 py-4 sm:py-6 lg:py-8">
        <NotFoundPage />
      </div>
    );
  }

  if (!profileData) {
    return null;
  }

  const { user, posts } = profileData.data;
  const displayName = user.first_name && user.last_name 
    ? `${user.first_name} ${user.last_name}` 
    : user.username;

  return (
    <div className="container mx-auto px-3 sm:px-4 lg:px-6 py-3 sm:py-4 lg:py-6">
      {/* User Profile Section */}
      <Card className="mb-4 sm:mb-6 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 border border-slate-700/50 shadow-lg relative">
        {/* Edit Profile Button */}
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
                  {/* Profile Picture Section */}
                  <div className="text-center space-y-4">
                    <div className="flex flex-col items-center space-y-3">
                      <Avatar className="h-20 w-20 border-4 border-orange-500/50 shadow-lg">
                        <AvatarImage src={imagePreview || editForm.avatar || user.avatar} alt="Profile" />
                        <AvatarFallback className="text-2xl bg-gradient-to-br from-orange-500 to-red-500 text-white">
                          {displayName.charAt(0).toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      
                      {/* Upload Progress */}
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
                              // Validate file using image service
                              const validation = imageService.validateFile(file);
                              if (!validation.isValid) {
                                warning(validation.error || 'Invalid file format');
                                return;
                              }
                              
                              setSelectedImageFile(file);
                              const previewUrl = imageService.createPreviewUrl(file);
                              setImagePreview(previewUrl);
                              
                              // Upload image immediately
                              setIsUploading(true);
                              setUploadProgress(0);
                              
                              try {
                                const uploadedUrl = await imageService.uploadProfileImage(file, {
                                  onProgress: (event) => {
                                    setUploadProgress(event.progress);
                                  },
                                  onSuccess: (url) => {
                                  },
                                  onError: (error) => {
                                    warning(`Failed to upload image: ${error.message}`);
                                  }
                                });
                                
                                // Store uploaded image data
                                setUploadedImageData({
                                  url: uploadedUrl,
                                  id: uploadedUrl.split('/').pop() || uploadedUrl // Extract filename as ID
                                });
                                
                              } catch (error) {
                                warning('Failed to upload image. Please try again.');
                                // Reset states on error
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
                        <p className="text-xs text-slate-400">
                          Click to upload new profile picture
                        </p>
                        
                        {/* Remove Image Button */}
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

                  {/* Basic Info */}
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

                  {/* Location & Website */}
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
                          onChange={(e) => setEditForm(prev => ({ ...prev, website: e.target.value }))}
                          placeholder="yourwebsite.com"
                          className="h-9 text-sm"
                        />
                      </div>
                    </div>
                  </div>

                  {/* Social Media */}
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
                          onChange={(e) => setEditForm(prev => ({ ...prev, github: e.target.value }))}
                          placeholder="username or URL"
                          className="h-9 text-sm"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="twitter" className="flex items-center gap-2 text-sm">
                          <Twitter className="w-4 h-4" />
                          Twitter/X
                        </Label>
                        <Input
                          id="twitter"
                          value={editForm.twitter}
                          onChange={(e) => setEditForm(prev => ({ ...prev, twitter: e.target.value }))}
                          placeholder="username or URL"
                          className="h-9 text-sm"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="linkedin" className="flex items-center gap-2 text-sm">
                          <Linkedin className="w-4 h-4" />
                          LinkedIn
                        </Label>
                        <Input
                          id="linkedin"
                          value={editForm.linkedin}
                          onChange={(e) => setEditForm(prev => ({ ...prev, linkedin: e.target.value }))}
                          placeholder="username or URL"
                          className="h-9 text-sm"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="instagram" className="flex items-center gap-2 text-sm">
                          <Instagram className="w-4 h-4" />
                          Instagram
                        </Label>
                        <Input
                          id="instagram"
                          value={editForm.instagram}
                          onChange={(e) => setEditForm(prev => ({ ...prev, instagram: e.target.value }))}
                          placeholder="username or URL"
                          className="h-9 text-sm"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="facebook" className="flex items-center gap-2 text-sm">
                          <Facebook className="w-4 h-4" />
                          Facebook
                        </Label>
                        <Input
                          id="facebook"
                          value={editForm.facebook}
                          onChange={(e) => setEditForm(prev => ({ ...prev, facebook: e.target.value }))}
                          placeholder="username or URL"
                          className="h-9 text-sm"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="youtube" className="flex items-center gap-2 text-sm">
                          <Youtube className="w-4 h-4" />
                          YouTube
                        </Label>
                        <Input
                          id="youtube"
                          value={editForm.youtube}
                          onChange={(e) => setEditForm(prev => ({ ...prev, youtube: e.target.value }))}
                          placeholder="channel name or URL"
                          className="h-9 text-sm"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="discord" className="flex items-center gap-2 text-sm">
                          <MessageCircle className="w-4 h-4" />
                          Discord
                        </Label>
                        <Input
                          id="discord"
                          value={editForm.discord}
                          onChange={(e) => setEditForm(prev => ({ ...prev, discord: e.target.value }))}
                          placeholder="username#1234"
                          className="h-9 text-sm"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="telegram" className="flex items-center gap-2 text-sm">
                          <Send className="w-4 h-4" />
                          Telegram
                        </Label>
                        <Input
                          id="telegram"
                          value={editForm.telegram}
                          onChange={(e) => setEditForm(prev => ({ ...prev, telegram: e.target.value }))}
                          placeholder="username"
                          className="h-9 text-sm"
                        />
                      </div>
                    </div>
                  </div>

                  {/* Action Buttons */}
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
            {/* Avatar */}
            <Avatar className="h-12 w-12 sm:h-16 sm:w-16 md:h-20 md:w-20 lg:h-24 lg:w-24 border-3 border-orange-500/50 shadow-lg">
              <AvatarImage src={user.avatar} alt={displayName} />
              <AvatarFallback className="text-lg sm:text-xl lg:text-2xl bg-gradient-to-br from-orange-500 to-red-500 text-white">
                {displayName.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>

            {/* Name and Handle */}
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

            {/* Bio */}
            {user.bio && (
              <p className="text-slate-300 max-w-lg text-center text-xs sm:text-sm leading-relaxed px-2">
                {user.bio}
              </p>
            )}

            {/* Location, Website, Joined */}
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

            {/* Social Media Links */}
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

            {/* Social Stats */}
            {/* <div className="flex items-center gap-2 sm:gap-3 text-center">
              <div>
                <div className="text-base sm:text-lg font-bold text-white">
                  {user.followers?.toLocaleString() || '0'}
                </div>
                <div className="text-xs text-slate-400">followers</div>
              </div>
              <div>
                <div className="text-base sm:text-lg font-bold text-white">
                  {user.following?.toLocaleString() || '0'}
                </div>
                <div className="text-xs text-slate-400">following</div>
              </div>
            </div> */}

            {/* Action Buttons */}
            {/* <div className="flex flex-col sm:flex-row items-center gap-1.5 sm:gap-2 w-full sm:w-auto">
              <Button className="w-full sm:w-auto px-3 sm:px-4 py-1.5 bg-orange-600 hover:bg-orange-700 text-white font-medium border-orange-500/30 text-xs">
                Follow
              </Button>
              <Button variant="outline" className="w-full sm:w-auto px-3 sm:px-4 py-1.5 border-orange-500/30 text-orange-300 hover:bg-orange-500/20 text-xs">
                @{user.username}
              </Button>
            </div> */}
            
          </div>
        </CardHeader>
      </Card>

      {/* Latest Posts Section */}
      <div className="mb-3 sm:mb-4">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1.5 sm:gap-2 mb-3 sm:mb-4">
          <h2 className="text-lg sm:text-xl font-bold text-white">Latest Posts</h2>
          <span className="text-xs sm:text-sm text-slate-400">
            {posts.meta.total} articles
          </span>
        </div>

        {posts.posts.length === 0 ? (
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
            {posts.posts.map((post) => (
              <BlogCard key={post.slug} post={post} />
            ))}
          </div>
        )}

        {/* Load More Button */}
        {posts.meta.hasNextPage && (
          <div className="text-center mt-4 sm:mt-6">
            <Button
              onClick={loadMorePosts}
              variant="outline"
              className="px-3 sm:px-4 py-1.5 text-xs sm:text-sm border-orange-500/30 text-orange-300 hover:bg-orange-500/20 w-full sm:w-auto"
            >
              Load More Posts
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
