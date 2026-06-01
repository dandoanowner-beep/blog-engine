# User Stories — Blog Engine
# Version: 1.0 — 2026-05-30

---

## Epic 1: Authentication & Account

### US-001: Email Registration
```gherkin
Feature: Email Registration
  As a visitor
  I want to register with my email and password
  So that I can become a member of the blog platform

  Scenario: Successful registration
    Given I am a guest on the platform
    When I fill in a valid email, username, and password
    And I submit the registration form
    Then my account is created with status "unverified"
    And I receive a verification email
    And I am redirected to a "Please verify your email" page

  Scenario: Registration with duplicate email
    Given a user with email "test@example.com" already exists
    When I try to register with "test@example.com"
    Then I see an error "Email already in use"
    And no new account is created

  Scenario: Registration with invalid password
    Given I am on the registration page
    When I submit a password shorter than 8 characters
    Then I see a validation error
    And no account is created
```

### US-002: Google OAuth Registration & Login
```gherkin
Feature: Google OAuth
  As a visitor
  I want to sign up or log in with my Google account
  So that I don't need to manage a separate password

  Scenario: First-time Google sign-up
    Given I am a guest
    When I click "Sign in with Google" and complete Google OAuth
    Then a new verified account is created using my Google email and name
    And I am logged in immediately
    And I can publish blogs without email verification

  Scenario: Returning Google user login
    Given I have previously signed up with Google
    When I click "Sign in with Google" and complete OAuth
    Then I am logged in to my existing account
```

### US-003: Email Verification
```gherkin
Feature: Email Verification
  As a registered user
  I want to verify my email
  So that I can publish blogs

  Scenario: Successful verification
    Given I registered with email and received a verification link
    When I click the verification link within 24 hours
    Then my account status is set to "verified"
    And I can now publish blogs

  Scenario: Expired verification link
    Given my verification link is older than 24 hours
    When I click it
    Then I see "Link expired"
    And I am prompted to request a new verification email

  Scenario: Unverified user tries to publish
    Given I am logged in but unverified
    When I try to publish a blog
    Then I see "Please verify your email to publish"
    And the blog is not published
```

### US-004: Login
```gherkin
Feature: Login
  As a registered user
  I want to log in
  So that I can access my account

  Scenario: Successful email login
    Given I have a verified account
    When I enter correct email and password
    Then I am logged in and redirected to the Explore feed

  Scenario: Wrong password
    Given I have an account
    When I enter the wrong password
    Then I see "Invalid email or password"
    And I remain on the login page

  Scenario: Password reset request
    Given I forgot my password
    When I click "Forgot password" and enter my email
    Then I receive a password reset email within 1 minute
```

### US-005: Block User
```gherkin
Feature: Block User
  As a signed-in user
  I want to block another user
  So that we cannot see each other's content

  Scenario: Blocking a user
    Given I am viewing user Bob's profile
    When I click "Block Bob"
    Then Bob is added to my blocked list
    And Bob's blogs no longer appear in any of my feeds
    And my blogs no longer appear in Bob's feeds
    And Bob cannot view my profile

  Scenario: Blocked user tries to view my profile
    Given Bob has been blocked by me
    When Bob navigates to my profile URL
    Then Bob sees "User not found" or a blank profile
```

---

## Epic 2: Blog Management

### US-006: Create and Publish Blog
```gherkin
Feature: Create Blog
  As a verified user
  I want to create and publish a blog
  So that others can read my content

  Scenario: Successful blog creation
    Given I am a verified logged-in user
    When I open the blog editor
    And I add a title, content (with WYSIWYG), thumbnail image, at least one tag, and one category
    And I select privacy "Public"
    And I click Publish
    Then the blog is published and appears in the Explore feed
    And other users can see it

  Scenario: Missing required fields
    Given I am in the blog editor
    When I try to publish without a title or without a tag
    Then I see validation errors for missing fields
    And the blog is not published

  Scenario: Image upload exceeds 5MB
    Given I am in the blog editor
    When I try to upload an image larger than 5MB
    Then I see "Image must be under 5MB"
    And the image is not uploaded
```

### US-007: Save Draft
```gherkin
Feature: Blog Draft
  As a writer
  I want to save my blog as a draft
  So that I can finish it later

  Scenario: Save draft
    Given I am writing a blog
    When I click "Save Draft"
    Then the blog is saved with status "draft"
    And it is not visible to anyone else
    And I can find it in "My Drafts"

  Scenario: Publish a draft
    Given I have a saved draft
    When I open it, make final edits, and click Publish
    Then the blog status changes to "published"
    And it appears in the Explore feed according to its privacy setting
```

### US-008: Privacy Modes
```gherkin
Feature: Blog Privacy
  As a writer
  I want to set the privacy of my blog
  So that I control who can see it

  Scenario: Public blog visible to guest
    Given I published a blog with privacy "Public"
    When a guest opens my blog card
    Then they can read the top 30% of the content
    And the rest is hidden with a signup prompt

  Scenario: Friend-only blog hidden from non-friends
    Given I published a blog with privacy "Friend-only"
    When a user who is NOT my friend visits my profile
    Then they cannot see that blog in my profile grid
    And the blog does not appear in their Explore or Following feed

  Scenario: Only-me blog invisible to everyone
    Given I published a blog with privacy "Only-me"
    When any other user (including Admin) visits my profile
    Then they cannot see that blog
    And it does not appear in any feed or search result
```

### US-009: Delete Blog
```gherkin
Feature: Delete Blog
  As a writer
  I want to delete my blog
  So that it is permanently removed

  Scenario: Writer deletes own blog
    Given I am the author of a published blog
    When I click "Delete" on my blog
    And I confirm the deletion
    Then the blog is permanently removed
    And all comments and reactions on it are also removed

  Scenario: Moderator deletes any blog
    Given I am a Moderator
    When I navigate to any blog and click "Delete"
    Then the blog is permanently removed regardless of author
```

---

## Epic 3: Feed & Discovery

### US-010: Explore Feed
```gherkin
Feature: Explore Feed
  As a signed-in user
  I want to browse the Explore feed
  So that I can discover interesting public blogs

  Scenario: View Explore feed
    Given I am logged in
    When I open the Explore tab
    Then I see public blogs displayed as cards, 3 per row
    And blogs are ranked by the algorithm (recency + engagement + followed writers boosted)
    And results are paginated with numbered pages

  Scenario: Filter Explore by tag
    Given I am on the Explore tab
    When I click the tag "technology"
    Then only blogs tagged "technology" are shown
```

### US-011: Following Feed
```gherkin
Feature: Following Feed
  As a signed-in user
  I want to see blogs from people I follow
  So that I don't miss their content

  Scenario: View Following feed
    Given I follow users Alice and Bob
    When I open the Following tab
    Then I see only blogs published by Alice and Bob
    And blogs are ordered newest first
    And results are paginated
```

---

## Epic 4: Social Features

### US-012: Follow User
```gherkin
Feature: Follow User
  As a signed-in user
  I want to follow other writers
  So that their content appears in my Following feed

  Scenario: Follow a user
    Given I am on Alice's profile page
    When I click "Follow"
    Then I am following Alice
    And Alice receives a notification "Someone followed you"
    And Alice's blogs now appear in my Following tab

  Scenario: Unfollow a user
    Given I am following Alice
    When I click "Unfollow" on her profile
    Then I no longer follow Alice
    And her blogs are removed from my Following tab
```

### US-013: Friend Request
```gherkin
Feature: Friend System
  As a signed-in user
  I want to send and manage friend requests
  So that I can share friend-only content with close connections

  Scenario: Send friend request
    Given I am on Bob's profile
    When I click "Add Friend"
    Then Bob receives a notification "Someone sent you a friend request"
    And my request appears as "Pending" on my end

  Scenario: Accept friend request
    Given Bob sent me a friend request
    When I accept it
    Then Bob and I become mutual friends
    And Bob receives a notification "Your friend request was accepted"
    And I can now see Bob's friend-only blogs and vice versa

  Scenario: Reject friend request
    Given Bob sent me a friend request
    When I reject it
    Then the request is removed
    And Bob is NOT notified of the rejection
```

### US-014: Like and Dislike
```gherkin
Feature: Reactions
  As a signed-in user
  I want to like or dislike a blog
  So that I can express my reaction

  Scenario: Like a blog
    Given I am viewing a blog
    When I click the Like button
    Then the like count increases by 1
    And the blog author receives a notification "Someone liked your blog"
    And my like is highlighted

  Scenario: Switch from like to dislike
    Given I have already liked a blog
    When I click the Dislike button
    Then my like is removed
    And a dislike is added
    And the counts update accordingly

  Scenario: Remove reaction
    Given I have liked a blog
    When I click Like again
    Then my like is removed
    And the like count decreases by 1
```

### US-015: Threaded Comments
```gherkin
Feature: Comments
  As a signed-in user
  I want to comment on blogs and reply to comments
  So that I can engage in discussion

  Scenario: Post a comment
    Given I am reading a blog
    When I type a comment and submit
    Then my comment appears at the bottom of the blog
    And the blog author receives a notification "Someone commented on your blog"
    And the comment count on the blog card increases by 1

  Scenario: Reply to a comment
    Given there is a comment by Alice on a blog
    When I click "Reply" on Alice's comment and submit my reply
    Then my reply appears nested under Alice's comment
    And Alice receives a notification "Someone replied to your comment"

  Scenario: Delete own comment
    Given I posted a comment
    When I click Delete on my comment
    Then the comment is permanently removed
    And all replies to it are also removed
```

### US-016: Report Content
```gherkin
Feature: Report System
  As a signed-in user
  I want to report inappropriate blogs or comments
  So that Moderators can review and take action

  Scenario: Report a blog
    Given I am reading a blog I find inappropriate
    When I click "Report" and select a reason
    Then the report is submitted silently
    And I see "Report submitted"
    And the blog author is NOT notified
    And all Moderators and Admins receive a notification "A blog has been reported"

  Scenario: Report a comment
    Given I see an inappropriate comment
    When I click "Report" on that comment and select a reason
    Then the report is submitted
    And all Moderators and Admins are notified
```

---

## Epic 5: Search

### US-017: Search
```gherkin
Feature: Universal Search
  As a signed-in user
  I want to search for blogs, users, and tags
  So that I can find specific content quickly

  Scenario: Search by blog title
    Given I type "Go tutorial" in the search bar
    When results load
    Then I see blogs whose title contains "Go tutorial"
    And results are grouped: Blogs / Users / Tags

  Scenario: Search respects privacy
    Given a blog has privacy "Friend-only" and I am not the author's friend
    When I search for that blog's exact title
    Then that blog does NOT appear in my search results

  Scenario: Search by username
    Given I type "alice" in the search bar
    When results load
    Then I see users whose name or username matches "alice"
```

---

## Epic 6: Admin Dashboard

### US-018: Admin User Management
```gherkin
Feature: Admin Dashboard
  As an Admin or Owner
  I want to manage users and reports
  So that I can maintain platform health

  Scenario: Promote user to Moderator
    Given I am an Admin viewing a user's profile in the dashboard
    When I click "Promote to Moderator"
    Then that user's role changes to Moderator
    And they can now delete blogs and receive report notifications

  Scenario: Review and act on a report
    Given a blog has been reported
    When I open the Reports Queue in the dashboard
    Then I see the reported blog, the reason, and the reporter
    When I click "Delete Blog"
    Then the blog is permanently removed
    And the report is marked as resolved
```
