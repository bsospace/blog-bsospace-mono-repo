import { axiosInstance } from '../utils/api';

export interface UploadProgress {
  progress: number;
}

export interface UploadOptions {
  maxSize?: number;
  onProgress?: (event: UploadProgress) => void;
  onSuccess?: (url: string) => void;
  onError?: (error: Error) => void;
}

export class ImageService {
  private static instance: ImageService;
  private maxSize: number = 5 * 1024 * 1024; // 5MB

  private constructor() {}

  public static getInstance(): ImageService {
    if (!ImageService.instance) {
      ImageService.instance = new ImageService();
    }
    return ImageService.instance;
  }

  /**
   * Upload profile image
   */
  async uploadProfileImage(
    file: File,
    options: UploadOptions = {}
  ): Promise<string> {
    // Validate file size
    if (file.size > this.maxSize) {
      const error = new Error(
        `File size exceeds maximum allowed (${this.maxSize / 1024 / 1024}MB)`
      );
      options.onError?.(error);
      throw error;
    }

    // Validate file type
    const validImageTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    if (!validImageTypes.includes(file.type)) {
      const error = new Error('Invalid file type. Only JPEG, PNG, GIF, and WebP are allowed.');
      options.onError?.(error);
      throw error;
    }

    const formData = new FormData();
    formData.append('file', file); // Changed from 'image' to 'file' to match media service

    try {
      const response = await axiosInstance.post('/media/upload', formData, { // Changed from '/image/upload' to '/media/upload'
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (event) => {
          if (event.total) {
            const percent = Math.round((event.loaded * 100) / event.total);
            options.onProgress?.({ progress: percent });
          }
        },
        withCredentials: true,
      });

      // Extract URL from media service response
      const url = response?.data?.data?.image_url || response?.data?.data?.url || response?.data?.data?.filename;

      if (!url) {
        throw new Error('No image URL returned from server');
      }

      options.onSuccess?.(url);
      return url;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Upload failed';
      const uploadError = new Error(`Image upload failed: ${errorMessage}`);
      options.onError?.(uploadError);
      throw uploadError;
    }
  }

  /**
   * Delete profile image
   */
  async deleteProfileImage(filename: string): Promise<boolean> {
    try {
      await axiosInstance.delete(`/image/${filename}`);
      return true;
    } catch (error) {
      console.error('Failed to delete image:', error);
      return false;
    }
  }

  /**
   * Get image info
   */
  async getImageInfo(filename: string): Promise<any> {
    try {
      const response = await axiosInstance.get(`/image/${filename}/info`);
      return response.data.data;
    } catch (error) {
      console.error('Failed to get image info:', error);
      return null;
    }
  }

  /**
   * Validate file before upload
   */
  validateFile(file: File): { isValid: boolean; error?: string } {
    // Check file size
    if (file.size > this.maxSize) {
      return {
        isValid: false,
        error: `File size exceeds maximum allowed (${this.maxSize / 1024 / 1024}MB)`
      };
    }

    // Check file type
    const validImageTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    if (!validImageTypes.includes(file.type)) {
      return {
        isValid: false,
        error: 'Invalid file type. Only JPEG, PNG, GIF, and WebP are allowed.'
      };
    }

    return { isValid: true };
  }

  /**
   * Create preview URL for file
   */
  createPreviewUrl(file: File): string {
    return URL.createObjectURL(file);
  }

  /**
   * Clean up preview URL
   */
  cleanupPreviewUrl(url: string): void {
    URL.revokeObjectURL(url);
  }

  /**
   * Format file size
   */
  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  }
}

// Export singleton instance
export const imageService = ImageService.getInstance();
