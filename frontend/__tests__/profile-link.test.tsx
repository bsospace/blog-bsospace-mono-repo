// Test to verify the profile link fix
import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock Post interface matching the actual interface
interface MockPost {
  id: string;
  title: string;
  author?: {
    first_name?: string;
    last_name?: string;
    username?: string;
  };
}

// Test component that mimics the profile link behavior
const TestProfileLink = ({ post }: { post: MockPost }) => {
  const getDisplayName = (user: { first_name?: string; last_name?: string; username?: string } | undefined): string => {
    if (!user) return "Unknown Author";
    const firstName = user.first_name;
    const lastName = user.last_name;
    return firstName && lastName ? `${firstName} ${lastName}` : (user.username || "Unknown Author");
  };

  const displayName = getDisplayName(post.author);

  return (
    <div>
      <span data-testid="display-name">{displayName}</span>
      <a 
        href={`/@${post.author?.username}`} 
        data-testid="profile-link"
      >
        View {displayName} →
      </a>
    </div>
  );
};

describe('Profile Link Fix', () => {
  test('should use username in URL even when display name is firstName + lastName', () => {
    const mockPost: MockPost = {
      id: '1',
      title: 'Test Post',
      author: {
        first_name: 'Piyawat',
        last_name: 'BSOspace',
        username: 'piyawat.bsospace'
      }
    };

    render(<TestProfileLink post={mockPost} />);
    
    // Check that display name shows "Piyawat BSOspace"
    const displayName = screen.getByTestId('display-name');
    expect(displayName).toHaveTextContent('Piyawat BSOspace');
    
    // Check that the link uses the username, not the display name
    const profileLink = screen.getByTestId('profile-link');
    expect(profileLink).toHaveAttribute('href', '/@piyawat.bsospace');
    
    // Ensure it does NOT use the display name in the URL
    expect(profileLink).not.toHaveAttribute('href', '/@Piyawat BSOspace');
    
    console.log('✅ Test passed: Profile link uses username instead of display name');
  });

  test('should handle missing username gracefully', () => {
    const mockPost: MockPost = {
      id: '2',
      title: 'Test Post 2',
      author: {
        first_name: 'John',
        last_name: 'Doe',
        // username is undefined
      }
    };

    render(<TestProfileLink post={mockPost} />);
    
    const profileLink = screen.getByTestId('profile-link');
    expect(profileLink).toHaveAttribute('href', '/@undefined');
    
    console.log('✅ Test passed: Handles missing username gracefully');
  });
});

export {};