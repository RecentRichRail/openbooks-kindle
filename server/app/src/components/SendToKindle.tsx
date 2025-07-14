import {
  Button,
  Loader,
  Modal,
  Stack,
  Text,
  TextInput,
  Group
} from "@mantine/core";
import { useState, useEffect } from "react";
import { useAppDispatch, useAppSelector } from "../state/store";
import { sendToKindle, sendDownload } from "../state/stateSlice";
import { MessageType } from "../state/messages";

interface SendToKindleProps {
  book: string;
  title: string;
  author: string;
}

// Cookie utility functions
const KINDLE_EMAIL_COOKIE = 'openbooks_kindle_email';

const setCookie = (name: string, value: string, days: number = 365) => {
  const expires = new Date();
  expires.setTime(expires.getTime() + (days * 24 * 60 * 60 * 1000));
  document.cookie = `${name}=${value};expires=${expires.toUTCString()};path=/`;
};

const getCookie = (name: string): string | null => {
  const nameEQ = name + "=";
  const ca = document.cookie.split(';');
  for (let i = 0; i < ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) === ' ') c = c.substring(1, c.length);
    if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
  }
  return null;
};

export default function SendToKindle({ book, title, author }: SendToKindleProps) {
  const [opened, setOpened] = useState(false);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [email, setEmail] = useState('');
  const [emailError, setEmailError] = useState<string | null>(null);
  const [status, setStatus] = useState<string>('');
  const [downloadComplete, setDownloadComplete] = useState(false);
  
  const dispatch = useAppDispatch();
  const notifications = useAppSelector(state => state.notifications.notifications);

  // Listen for WebSocket notifications
  useEffect(() => {
    if (!loading) return;
    
    const latestNotification = notifications[0];
    if (!latestNotification) return;

    console.log('Notification received:', latestNotification.title, 'Appearance:', latestNotification.appearance);

    // Skip informational notifications that we don't want to act on
    if (latestNotification.title.includes('Download request sent') || 
        latestNotification.title.includes('Waiting for book to download')) {
      console.log('Ignoring informational notification');
      return;
    }
    
    // Listen for download completion (specifically the "Book downloaded!" message)
    if (!downloadComplete && latestNotification.title.includes('Book downloaded!')) {
      console.log('Download completion detected');
      setDownloadComplete(true);
      setStatus('Sending book to your Kindle email...');
      return;
    }

    // Listen for send to kindle completion (exact match)
    if (latestNotification.title === 'Book sent to your email successfully!') {
      console.log('Email success detected');
      setStatus('Book sent successfully!');
      setSuccess(true);
      setLoading(false);
      
      // Auto-close after 3 seconds on success
      setTimeout(() => {
        setOpened(false);
        setSuccess(false);
        setStatus('');
        setDownloadComplete(false);
      }, 3000);
      return;
    }

    // Only treat DANGER type notifications as errors
    if (latestNotification.appearance === 2) { // DANGER type only
      console.log('Error notification detected');
      setError('Failed to send book to Kindle. Please try again.');
      setStatus('');
      setLoading(false);
      setDownloadComplete(false);
      return;
    }
    
    // Log any other notifications we're ignoring
    console.log('Ignoring notification during loading phase');
  }, [notifications, loading, downloadComplete]);

  // Load saved email from cookie on component mount
  useEffect(() => {
    const savedEmail = getCookie(KINDLE_EMAIL_COOKIE);
    if (savedEmail) {
      setEmail(savedEmail);
    }
  }, []);

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^\S+@\S+\.\S+$/;
    if (!emailRegex.test(email)) {
      setEmailError('Invalid email address');
      return false;
    }
    setEmailError(null);
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateEmail(email)) {
      return;
    }
    
    // Save email to cookie for future use
    setCookie(KINDLE_EMAIL_COOKIE, email);
    
    setLoading(true);
    setError(null);
    setSuccess(false);
    setDownloadComplete(false);
    
    try {
      // Step 1: Download the book
      setStatus('Downloading book from IRC...');
      dispatch(sendDownload(book));
      
      // Step 2: Send email (will be triggered after download completes via notification listener)
      dispatch(sendToKindle({
        book: book,
        email,
        title: title,
        author: author
      }));
      
      // The rest will be handled by the notification listener useEffect
    } catch (err) {
      setError('Failed to send book to Kindle. Please try again.');
      setStatus('');
      setLoading(false);
      setDownloadComplete(false);
    }
  };

  const handleClose = () => {
    // Prevent closing during loading
    if (loading) {
      return;
    }
    
    setOpened(false);
    setLoading(false);
    setSuccess(false);
    setError(null);
    setEmailError(null);
    setStatus('');
    setDownloadComplete(false);
    // Don't clear email on close - keep it for convenience
  };

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    if (emailError) {
      setEmailError(null);
    }
  };

  return (
    <>
      <Button
        compact
        size="xs"
        radius="sm"
        onClick={() => setOpened(true)}
        sx={{ fontWeight: "normal", width: 100 }}
        color="green"
      >
        Request
      </Button>

      <Modal
        opened={opened}
        onClose={handleClose}
        title={`Request "${author}" for Kindle`}
        centered
        closeOnClickOutside={!loading}
        closeOnEscape={!loading}
        withCloseButton={!loading}
      >
        {loading ? (
          <Stack align="center" spacing="md">
            <Loader size="lg" />
            <Text weight={500} align="center">
              {status}
            </Text>
            <Text size="sm" color="dimmed" align="center">
              Please wait while we process your request...
            </Text>
          </Stack>
        ) : success ? (
          <Stack align="center" spacing="md">
            <Text color="green" weight={500}>
              ✓ {status}
            </Text>
            <Text size="sm" color="dimmed" align="center">
              The book has been downloaded and sent to your Kindle email address.
            </Text>
          </Stack>
        ) : (
          <form onSubmit={handleSubmit}>
            <Stack spacing="md">
              <Text size="sm" color="dimmed">
                Enter your Kindle email address. We'll download "{author}" by {title} and send it to your Kindle device.
              </Text>
              
              <TextInput
                label="Kindle Email Address"
                placeholder="your-kindle@kindle.com"
                required
                value={email}
                onChange={handleEmailChange}
                error={emailError}
                disabled={loading}
                description="Your email will be saved for future requests"
              />

              {error && (
                <Text color="red" size="sm">
                  {error}
                </Text>
              )}

              <Group position="right" spacing="sm">
                <Button variant="subtle" onClick={handleClose} disabled={loading}>
                  Cancel
                </Button>
                <Button 
                  type="submit" 
                  loading={loading}
                  disabled={loading}
                >
                  Send to Kindle
                </Button>
              </Group>
            </Stack>
          </form>
        )}
      </Modal>
    </>
  );
}
