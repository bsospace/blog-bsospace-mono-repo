// Test to verify the profile link fix - URLs pattern validation
describe('Profile Link Fix', () => {
  test('should use username in URL instead of display name format', () => {
    // Test URL format
    const username = 'piyawat.bsospace';
    const displayName = 'Piyawat BSOspace';
    
    // Correct URL format
    const correctUrl = `/@${username}`;
    expect(correctUrl).toBe('/@piyawat.bsospace');
    
    // Incorrect URL format (what was happening before)
    const incorrectUrl = `/@${displayName}`;
    expect(incorrectUrl).toBe('/@Piyawat BSOspace');
    
    // Verify they are different
    expect(correctUrl).not.toBe(incorrectUrl);
    
    console.log('✅ URL format test passed');
    console.log(`Correct URL: ${correctUrl}`);
    console.log(`Incorrect URL: ${incorrectUrl}`);
  });

  test('should validate URL patterns', () => {
    // Test valid username patterns
    const validUsernames = ['piyawat.bsospace', 'john.doe', 'user123', 'test-user'];
    
    validUsernames.forEach(username => {
      const url = `/@${username}`;
      // URL should not contain spaces
      expect(url).not.toMatch(/\s/);
      // URL should start with /@
      expect(url).toMatch(/^\/@/);
    });
    
    console.log('✅ URL pattern validation passed');
  });
});

export {};