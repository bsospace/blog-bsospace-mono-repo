"use client";

import React, { FC } from "react";
import Link from "next/link";
import { Post } from "../interfaces";
import { FiBookmark, FiClock, FiEye, FiHeart, FiMessageCircle, FiCpu } from "react-icons/fi";
import { formatRelativeTime } from "@/lib/utils";
import Image from "next/image";

const BlogCard = ({ post }
  : {
    post: Post;
  } & React.HTMLAttributes<HTMLDivElement>
  
) => {

  const getImageHeight = () => { 
    if (post) {
      return `${!post.ai_ready ? "h-40 sm:h-48 md:h-[230px]" : "h-40 sm:h-48 md:h-[260px]"}`;
    }
  }

  return (
    <div className="group relative bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 rounded-xl overflow-hidden border border-slate-700/50 hover:border-orange-500/50 transition-all duration-300 hover:shadow-lg hover:shadow-orange-500/10">

      {/* Animated border effect */}
      <div className="absolute inset-0 bg-gradient-to-r from-orange-500/0 via-orange-500/10 to-red-500/0 opacity-0 group-hover:opacity-100 transition-opacity duration-300 rounded-xl"></div>

      {/* Glowing orb effect */}
      <div className="absolute -top-16 -right-16 w-32 h-32 bg-gradient-to-br from-orange-500/20 to-red-500/20 rounded-full blur-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>

      <Link href={`/posts/@${post.author?.username}/${post.slug}`} className="relative z-10 flex flex-col md:flex-row group-hover:scale-102 transition-transform duration-300">
        {/* Thumbnail Section */}
        <div className="relative w-full md:w-2/5 flex-shrink-0">
          <div className={`h-40 overflow-hidden` + " " + getImageHeight()}>
            <div className="relative w-full h-full">
              <Image
                src={post?.thumbnail || '/default-thumbnail.png'}
                alt={post.title}
                fill
                className="object-cover transform group-hover:scale-105 transition-transform duration-500"
                sizes="100%"
                priority
              />
            </div>
            {/* Tech overlay */}
            <div className="absolute inset-0 bg-gradient-to-tr from-black/50 via-transparent to-orange-500/20 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>

            {/* Floating tech elements */}
            <div className="absolute top-3 left-3 opacity-0 group-hover:opacity-100 transition-all duration-300 delay-100">
              <div className="flex space-x-1.5">
                <div className="w-1.5 h-1.5 bg-orange-400 rounded-full animate-pulse"></div>
                <div className="w-1.5 h-1.5 bg-red-400 rounded-full animate-pulse delay-100"></div>
                <div className="w-1.5 h-1.5 bg-yellow-400 rounded-full animate-pulse delay-200"></div>
              </div>
            </div>

            {/* AI Processing Badge */}
            {post.ai_chat_open && !post.ai_ready && (
              <div className="absolute bottom-3 left-3">
                <div className="flex items-center gap-1.5 bg-gradient-to-r from-yellow-600/90 to-orange-600/90 backdrop-blur-md px-2 py-1 rounded-full border border-yellow-400/30 shadow-lg">
                  <FiCpu className="w-2.5 h-2.5 text-yellow-300 animate-pulse" />
                  <span className="text-[10px] text-yellow-100">AI Processing</span>
                  <div className="w-1 h-1 bg-yellow-400 rounded-full animate-bounce"></div>
                </div>
              </div>
            )}
          </div>

          {/* Bookmark with tech styling */}
          <button className="absolute top-3 right-3 bg-black/50 backdrop-blur-md p-1.5 rounded-lg border border-orange-500/30 opacity-0 group-hover:opacity-100 transition-all duration-300 hover:bg-orange-500/20">
            <FiBookmark className="w-3.5 h-3.5 text-orange-400" />
          </button>
        </div>

        {/* Content Section */}
        <div className="w-full md:w-3/5 p-3 sm:p-4 flex flex-col min-h-0">
          {/* Tech Tags */}
          {post.tags && post.tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mb-2 sm:mb-3">
              {post.tags.slice(0, 3).map((tag) => (
                <span
                  key={tag.id}
                  className="px-2 py-0.5 text-xs font-mono bg-gradient-to-r from-orange-500/20 to-red-500/20 text-orange-300 rounded-full border border-orange-500/30 hover:border-orange-400 transition-colors cursor-pointer"
                >
                  #{tag.name}
                </span>
              ))}
            </div>
          )}

          {/* AI Mode Indicator in Content */}
          {post.ai_chat_open && (
            <div className="flex items-center gap-1 mb-2">
              <div className="flex items-center gap-1 px-1 py-0.5 rounded-md bg-gradient-to-r from-orange-500/10 to-red-500/10 border border-orange-500/20">
                <FiMessageCircle className="w-2.5 h-2.5 text-orange-400" />
                <span className="text-[10px] text-orange-300 font-medium">
                  {post.ai_ready ? "Interactive AI Chat Available" : "AI Processing Content"}
                </span>
              </div>
            </div>
          )}

          {/* Title with tech styling (no inner link to avoid nested anchors) */}
          <h2 className="flex w-full hover:underline hover:underline-offset-4 group/title mb-2 sm:mb-3 text-sm sm:text-base md:text-lg font-bold text-black dark:text-white line-clamp-2 group-hover/title:text-transparent group-hover/title:bg-gradient-to-r group-hover/title:from-orange-400 group-hover/title:to-red-400 group-hover/title:bg-clip-text transition-all duration-300">
            {post.title}
          </h2>

          {/* Description - จะยืดหยุ่นตามเนื้อหา */}
          {post.description && (
            <p className="text-slate-300 text-xs sm:text-sm mb-3 line-clamp-3 leading-relaxed flex-grow">
              {post.description}
            </p>
          )}

          {/* Bottom section - อยู่ด้านล่างเสมอ */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 sm:gap-3 mt-auto pt-3 border-t border-slate-700/50">

            {/* Author info */}
            <div className="flex items-center gap-2">
              <div className="relative">
                <div className="w-7 h-7 rounded-full overflow-hidden border-2 border-orange-600 shadow-sm">
                  <Image
                    src={post.author?.avatar || '/default-avatar.png'}
                    alt={post.author?.username || 'Author'}
                    width={28}
                    height={28}
                    className="w-full h-full object-cover"
                  />
                </div>
                <span className="absolute -bottom-0.5 -right-0.5 w-2 h-2 bg-green-400 rounded-full border-2 border-slate-900" />
              </div>

              <div className="leading-snug min-w-0 flex-1">
                <p className="text-xs font-medium m-0 p-0 truncate">
                  @{post.author?.username || "ผู้เขียน"}
                </p>
                <p className="text-slate-400 text-[10px] m-0 p-0">
                  {post.published_at ? formatRelativeTime(new Date(post.published_at)) : 'ไม่มีวันที่'}
                </p>
              </div>
            </div>

            {/* Stats */}
            <div className="flex items-center gap-2 sm:gap-4 flex-shrink-0">
              <div className="flex items-center gap-1 text-xs text-slate-400">
                <FiClock className="w-3 h-3 text-orange-400" />
                <span>{post.read_time || 0}m</span>
              </div>
              <div className="flex items-center gap-1 text-xs text-slate-400">
                <FiEye className="w-3 h-3 text-red-400" />
                <span>{(post.views || 0).toLocaleString()}</span>
              </div>
              <div className="flex items-center gap-1 text-xs text-slate-400">
                <FiHeart className="w-3 h-3 text-pink-400" />
                <span>{post.likes || 0}</span>
              </div>
              {/* AI Chat Indicator in Stats */}
              {post.ai_ready && post.ai_chat_open && (
                <div className="flex items-center gap-1 text-xs text-orange-400">
                  <FiCpu className="w-3 h-3" />
                  <span>AI</span>
                </div>
              )}
            </div>
          </div>

        </div>
      </Link>
    </div>
  );
};
export default BlogCard;