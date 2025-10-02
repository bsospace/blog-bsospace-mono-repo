/* eslint-disable @next/next/no-img-element */
"use client";
import { ReactNode, useContext, useEffect, useState, useRef } from "react";
import ThemeSwitcher from "./theme-switcher";
import Image from "next/image";
import logo from "../../public/logo.svg";
import Link from "next/link";
import axios from "axios";
import Script from "next/script";
import { AuthContext } from "../contexts/auth-context";
import { ChevronDown, LogOut, Notebook, SquarePen, User, UserCircle } from "lucide-react";
import { getnerateId } from "@/lib/utils";
import { FiCode, FiCpu } from "react-icons/fi";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import NotificationDropdown from "./notification-dropdown";
import { axiosInstance } from "../../lib/api";
import envConfig from '../configs/env-config';

// Profile Button Component
const ProfileButton = ({ 
  isLoggedIn, 
  user, 
  isOpen 
}: {
  isLoggedIn: boolean;
  user: any;
  isOpen: boolean;
}) => (
  <div className="flex items-center gap-2 p-1.5 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-all duration-200 group">
    {isLoggedIn && user ? (
      <>
        <div className="w-8 h-8 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow-sm ring-2 ring-orange-100 dark:ring-orange-900/20 group-hover:ring-orange-200 dark:group-hover:ring-orange-800 transition-all duration-200">
          <img src={user.avatar} alt="Profile" className="w-full h-full object-cover" />
        </div>
        <ChevronDown className={`w-4 h-4 text-gray-600 dark:text-gray-400 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} />
      </>
    ) : (
      <>
        <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gradient-to-br from-orange-100 to-orange-200 dark:from-orange-900/30 dark:to-orange-800/30 group-hover:from-orange-200 group-hover:to-orange-300 dark:group-hover:from-orange-800/50 dark:group-hover:to-orange-700/50 transition-all duration-200">
          <User className="w-5 h-5 text-orange-600 dark:text-orange-400" />
        </div>
        <ChevronDown className={`w-4 h-4 text-gray-600 dark:text-gray-400 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} />
      </>
    )}
  </div>
);

export default function Layout({ children }: { children: ReactNode }) {
  const [version, setVersion] = useState<string>("unknown");
  const [showNavbar, setShowNavbar] = useState(true);
  const { isLoggedIn, user, setIsLoggedIn, setUser } = useContext(AuthContext);
  const lastScrollY = useRef(0);

  const handleLogout = async () => {
    setIsLoggedIn(false);
    setUser(null);
    try {
      const response = await axiosInstance.delete("/auth/logout");
      if (response.data.success) {
        localStorage.removeItem("warp");
        localStorage.removeItem("logged_in");
        localStorage.removeItem("pid");
        window.location.href = "/home"
      }
    }
    catch (error) {
      console.error("Error during logout:", error);
    }
  };

  // Handle show/hide navbar on scroll (mobile only; always shown on md+)
  useEffect(() => {
    const onScroll = () => {
      const currentY = window.scrollY;
      if (currentY <= 16) {
        setShowNavbar(true);
      } else if (currentY > lastScrollY.current && currentY > 80) {
        setShowNavbar(false);
      } else {
        setShowNavbar(true);
      }
      lastScrollY.current = currentY;
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  // Add missing useEffect for version fetching
  useEffect(() => {
    const fetchLatestVersion = async () => {
      try {
        const response = await axios.get(
          "https://api.github.com/repos/bsospace/blog-bsospace-mono-repo/releases/latest"
        );
        setVersion(response.data.tag_name || "unknown");
      } catch (error) {
        console.error("Error fetching latest version from GitHub:", error);
        setVersion("unknown");
      }
    };

    fetchLatestVersion();
  }, []);

  return (
    <div className="flex flex-col min-h-screen dark:bg-space-dark bg-space-light">
      <Script
        data-name="BMC-Widget"
        data-cfasync="false"
        src="https://cdnjs.buymeacoffee.com/1.0.0/widget.prod.min.js"
        data-id="bsospace"
        data-description="Support me on Buy me a coffee!"
        data-message=""
        data-color="#FF813F"
        data-position="Right"
        data-x_margin="18"
        data-y_margin="18"
      ></Script>

      {/* Header: fixed, rounded, blurred navbar */}
      <div
        className={`fixed top-0 z-50 w-full transition-all duration-300 py-2 px-1 md:py-4 md:px-2 ${showNavbar ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-4"} md:opacity-100 md:translate-y-0`}
      >
        <div className="bg-white/70 dark:bg-gray-900/70 backdrop-blur-sm shadow-sm rounded-full max-w-6xl items-center mx-auto">
          <div className="mx-auto py-1 px-1 md:px-4 md:py-2 flex items-center">
            {/* Logo */}
            <Link href="/" className="flex items-center ms-4 md:ms-0 gap-2 no-underline text-black dark:text-white">
              <Image src={logo} alt="BSO logo" width={32} height={32} />
            </Link>

            {/* Right Controls */}
            <div className="w-full hidden md:flex justify-end items-center gap-4">
              {/* desktop only */}
              <nav className="hidden md:flex md:flex-end space-x-6 text-sm items-center font-medium text-gray-700 dark:text-gray-300">
                <a
                  href="https://github.com/bsospace"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-orange-500 no-underline transition-colors duration-200"
                >
                  GitHub
                </a>
                <a
                  href="https://www.youtube.com/@bsospace"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-orange-500 no-underline transition-colors duration-200"
                >
                  YouTube
                </a>
                <NotificationDropdown />
              </nav>
              {/* Separator */}
              <div className="hidden md:block border-l border-gray-200 dark:border-gray-700 h-6"></div>
              {/* Theme Switcher */}
              <ThemeSwitcher />
              {/* Profile Dropdown */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button className="flex items-center gap-2 p-1.5 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-orange-400 focus:ring-opacity-50 group">
                    {isLoggedIn && user ? (
                      <>
                        <div className="w-8 h-8 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow-sm ring-2 ring-orange-100 dark:ring-orange-900/20 group-hover:ring-orange-200 dark:group-hover:ring-orange-800 transition-all duration-200">
                          <img src={user.avatar} alt="Profile" className="w-full h-full object-cover" />
                        </div>
                        <ChevronDown className="w-4 h-4 text-gray-600 dark:text-gray-400 transition-transform duration-200" />
                      </>
                    ) : (
                      <>
                        <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gradient-to-br from-orange-100 to-orange-200 dark:from-orange-900/30 dark:to-orange-800/30 group-hover:from-orange-200 group-hover:to-orange-300 dark:group-hover:from-orange-800/50 dark:group-hover:to-orange-700/50 transition-all duration-200">
                          <User className="w-5 h-5 text-orange-600 dark:text-orange-400" />
                        </div>
                        <ChevronDown className="w-4 h-4 text-gray-600 dark:text-gray-400 transition-transform duration-200" />
                      </>
                    )}
                  </button>
                </DropdownMenuTrigger>
                
                <DropdownMenuContent 
                  className="w-72 p-0" 
                  align="end" 
                  sideOffset={8}
                  alignOffset={0}
                >
                  {isLoggedIn && user ? (
                    <>
                      {/* User Info Header */}
                      <div className="p-4 pb-3 border-b border-gray-200 dark:border-gray-700">
                        <div className="flex items-center gap-4">
                          <div className="w-12 h-12 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow-lg ring-2 ring-orange-100 dark:ring-orange-900/20">
                            <img src={user.avatar} alt="Profile" className="w-full h-full object-cover" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="font-semibold text-lg text-gray-900 dark:text-white truncate leading-tight">
                              {user.username || user.email.split("@")[0]}
                            </p>
                            <p className="text-sm text-gray-500 dark:text-gray-400 truncate leading-tight">{user.email}</p>
                          </div>
                        </div>
                      </div>

                      {/* Action Menu */}
                      <div className="p-2">
                        <DropdownMenuItem asChild>
                          <Link
                            href={`/w/${getnerateId()}`}
                            className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-orange-50 dark:hover:bg-orange-900/20 transition-all duration-200 text-gray-700 dark:text-gray-300 group cursor-pointer"
                          >
                            <SquarePen className="w-5 h-5 text-orange-500 group-hover:scale-110 transition-transform" />
                            <span className="font-medium">Write story</span>
                          </Link>
                        </DropdownMenuItem>
                        
                        <DropdownMenuItem asChild>
                          <Link
                            href="/w"
                            className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-all duration-200 text-gray-700 dark:text-gray-300 group cursor-pointer"
                          >
                            <Notebook className="w-5 h-5 text-blue-500 group-hover:scale-110 transition-transform" />
                            <span className="font-medium">My stories</span>
                          </Link>
                        </DropdownMenuItem>
                        
                        <DropdownMenuItem asChild>
                          <Link
                            href={`/@${user.username}`}
                            className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-green-50 dark:hover:bg-green-900/20 transition-all duration-200 text-gray-700 dark:text-gray-300 group cursor-pointer"
                          >
                            <UserCircle className="w-5 h-5 text-green-500 group-hover:scale-110 transition-transform" />
                            <span className="font-medium">Profile</span>
                          </Link>
                        </DropdownMenuItem>
                        
                        
                        <DropdownMenuSeparator />
                        
                        <DropdownMenuItem 
                          onClick={handleLogout}
                          className="flex items-center gap-3 px-3 py-2.5 text-sm text-red-600 dark:text-red-400 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-all duration-200 group cursor-pointer"
                        >
                          <LogOut className="w-5 h-5 group-hover:scale-110 transition-transform" />
                          <span className="font-medium">Sign out</span>
                        </DropdownMenuItem>
                      </div>
                    </>
                  ) : (
                    /* Guest User Menu */
                    <div className="p-4">
                      <div className="text-center text-gray-700 dark:text-gray-300 mb-4">
                        <h3 className="font-bold text-xl mb-4 bg-gradient-to-r from-orange-400 to-orange-600 text-transparent bg-clip-text">
                          Welcome
                        </h3>
                        
                        <div className="space-y-2 mb-4">
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/w/${getnerateId()}`}
                              className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-orange-50 dark:hover:bg-blue-900/20 transition-all duration-200 text-gray-700 dark:text-gray-300 group cursor-pointer"
                            >
                              <SquarePen className="w-5 h-5 text-orange-500 group-hover:scale-110 transition-transform" />
                              <span className="font-medium">Write story</span>
                            </Link>
                          </DropdownMenuItem>
                          
                          <DropdownMenuItem asChild>
                            <Link
                              href="/w"
                              className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-all duration-200 text-gray-400 dark:text-gray-600 cursor-not-allowed group"
                              onClick={(e) => e.preventDefault()}
                            >
                              <Notebook className="w-5 h-5 text-blue-400 group-hover:scale-110 transition-transform" />
                              <span className="font-medium">My stories</span>
                            </Link>
                          </DropdownMenuItem>
                          
                          <DropdownMenuItem asChild>
                            <Link
                              href=""
                              className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-green-50 dark:hover:bg-green-900/20 transition-all duration-200 text-gray-400 dark:text-gray-600 cursor-not-allowed group"
                              onClick={(e) => e.preventDefault()}
                            >
                              <UserCircle className="w-5 h-5 text-green-400 group-hover:scale-110 transition-transform" />
                              <span className="font-medium">Profile</span>
                            </Link>
                          </DropdownMenuItem>
                        </div>
                        
                        <Button 
                          variant="default" 
                          className="w-full bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white font-semibold py-2.5 rounded-lg transition-all duration-200 transform hover:scale-105" 
                          onClick={() => {
                            window.location.href = "/auth/login";
                          }}
                        >
                          Login
                        </Button>
                      </div>
                    </div>
                  )}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            {/* Mobile controls (Theme + Profile) */}
            <div className="md:hidden ml-auto flex items-center gap-2">
              <ThemeSwitcher />
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <button className="flex items-center gap-2 p-1.5 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-orange-400 focus:ring-opacity-50 group">
                    {isLoggedIn && user ? (
                      <>
                        <div className="w-7 h-7 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow-sm ring-2 ring-orange-100 dark:ring-orange-900/20 group-hover:ring-orange-200 dark:group-hover:ring-orange-800 transition-all duration-200">
                          <img src={user.avatar} alt="Profile" className="w-full h-full object-cover" />
                        </div>
                        <ChevronDown className="w-4 h-4 text-gray-600 dark:text-gray-400 transition-transform duration-200" />
                      </>
                    ) : (
                      <>
                        <div className="flex items-center justify-center w-7 h-7 rounded-full bg-gradient-to-br from-orange-100 to-orange-200 dark:from-orange-900/30 dark:to-orange-800/30 group-hover:from-orange-200 group-hover:to-orange-300 dark:group-hover:from-orange-800/50 dark:group-hover:to-orange-700/50 transition-all duration-200">
                          <User className="w-4 h-4 text-orange-600 dark:text-orange-400" />
                        </div>
                        <ChevronDown className="w-4 h-4 text-gray-600 dark:text-gray-400 transition-transform duration-200" />
                      </>
                    )}
                  </button>
                </DropdownMenuTrigger>
                
                <DropdownMenuContent 
                  className="w-72 p-0" 
                  align="end" 
                  sideOffset={8}
                  alignOffset={0}
                >
                  {isLoggedIn && user ? (
                    <>
                      {/* User Info Header */}
                      <div className="p-4 pb-3 border-b border-gray-200 dark:border-gray-700">
                        <div className="flex items-center gap-4">
                          <div className="w-12 h-12 rounded-full overflow-hidden border-2 border-white dark:border-gray-700 shadow-lg ring-2 ring-orange-100 dark:ring-orange-900/20">
                            <img src={user.avatar} alt="Profile" className="w-full h-full object-cover" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="font-semibold text-lg text-gray-900 dark:text-white truncate leading-tight">
                              {user.username || user.email.split("@")[0]}
                            </p>
                            <p className="text-sm text-gray-500 dark:text-gray-400 truncate leading-tight">{user.email}</p>
                          </div>
                        </div>
                      </div>

                      {/* Action Menu */}
                      <div className="p-2">
                        <DropdownMenuItem asChild>
                          <Link
                            href={`/w/${getnerateId()}`}
                            className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-orange-50 dark:hover:bg-orange-900/20 transition-all duration-200 text-gray-700 dark:text-gray-300 group cursor-pointer"
                          >
                            <SquarePen className="w-5 h-5 text-orange-500 group-hover:scale-110 transition-transform" />
                            <span className="font-medium">Write story</span>
                          </Link>
                        </DropdownMenuItem>
                        
                        <DropdownMenuItem asChild>
                          <Link
                            href="/w"
                            className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-all duration-200 text-gray-700 dark:text-gray-300 group cursor-pointer"
                          >
                            <Notebook className="w-5 h-5 text-blue-500 group-hover:scale-110 transition-transform" />
                            <span className="font-medium">My stories</span>
                          </Link>
                        </DropdownMenuItem>
                        
                        <DropdownMenuItem asChild>
                          <Link
                            href={`/@${user.username}`}
                            className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-green-50 dark:hover:bg-green-900/20 transition-all duration-200 text-gray-700 dark:text-gray-300 group cursor-pointer"
                          >
                            <UserCircle className="w-5 h-5 text-green-500 group-hover:scale-110 transition-transform" />
                            <span className="font-medium">Profile</span>
                          </Link>
                        </DropdownMenuItem>
                        
                        
                        <DropdownMenuSeparator />
                        
                        <DropdownMenuItem 
                          onClick={handleLogout}
                          className="flex items-center gap-3 px-3 py-2.5 text-sm text-red-600 dark:text-red-400 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-all duration-200 group cursor-pointer"
                        >
                          <LogOut className="w-5 h-5 group-hover:scale-110 transition-transform" />
                          <span className="font-medium">Sign out</span>
                        </DropdownMenuItem>
                      </div>
                    </>
                  ) : (
                    /* Guest User Menu */
                    <div className="p-4">
                      <div className="text-center text-gray-700 dark:text-gray-300 mb-4">
                        <h3 className="font-bold text-xl mb-4 bg-gradient-to-r from-orange-400 to-orange-600 text-transparent bg-clip-text">
                          Welcome
                        </h3>
                        
                        <div className="space-y-2 mb-4">
                          <DropdownMenuItem asChild>
                            <Link
                              href={`/w/${getnerateId()}`}
                              className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-orange-50 dark:hover:bg-blue-900/20 transition-all duration-200 text-gray-700 dark:text-gray-300 group cursor-pointer"
                            >
                              <SquarePen className="w-5 h-5 text-orange-500 group-hover:scale-110 transition-transform" />
                              <span className="font-medium">Write story</span>
                            </Link>
                          </DropdownMenuItem>
                          
                          <DropdownMenuItem asChild>
                            <Link
                              href="/w"
                              className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-all duration-200 text-gray-400 dark:text-gray-600 cursor-not-allowed group"
                              onClick={(e) => e.preventDefault()}
                            >
                              <Notebook className="w-5 h-5 text-blue-400 group-hover:scale-110 transition-transform" />
                              <span className="font-medium">My stories</span>
                            </Link>
                          </DropdownMenuItem>
                          
                          <DropdownMenuItem asChild>
                            <Link
                              href=""
                              className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-green-50 dark:hover:bg-green-900/20 transition-all duration-200 text-gray-400 dark:text-gray-600 cursor-not-allowed group"
                              onClick={(e) => e.preventDefault()}
                            >
                              <UserCircle className="w-5 h-5 text-green-400 group-hover:scale-110 transition-transform" />
                              <span className="font-medium">Profile</span>
                            </Link>
                          </DropdownMenuItem>
                        </div>
                        
                        <Button 
                          variant="default" 
                          className="w-full bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 text-white font-semibold py-2.5 rounded-lg transition-all duration-200 transform hover:scale-105" 
                          onClick={() => {
                            window.location.href = "/auth/login";
                          }}
                        >
                          Login
                        </Button>
                      </div>
                    </div>
                  )}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <main className="flex-grow w-full mx-auto md:p-6 p-4 pt-20 md:pt-24">

        {/* Background tech elements */}
        {children}
      </main>

      {/* Footer tech elements */}
      <footer className="text-center py-3 border-t border-slate-800">
        <div className="flex justify-center items-center space-x-4 text-slate-400">
          <FiCode className="w-5 h-5 text-orange-400" />
          <span className="md:text-sm text-[10px]">Be Simple but Outstanding | Version: {version} | &copy; {new Date().getFullYear()} <Link href="https://www.bsospace.com" target="_blank" className="hover:text-orange-400 transition-colors">{envConfig.organizationName}</Link></span>
          <FiCpu className="w-5 h-5 text-red-400" />
        </div>
      </footer>
    </div>
  );
}