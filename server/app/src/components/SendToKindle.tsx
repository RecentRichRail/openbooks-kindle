import {
  Button,
  Loader,
  Modal,
  Stack,
  Text,
  TextInput,
  Group
} from "@mantine/core";
import { useState } from "react";
import { useAppDispatch } from "../state/store";
import { sendToKindle } from "../state/stateSlice";

interface SendToKindleProps {
  book: string;
  title: string;
  author: string;
}

export default function SendToKindle({ book, title, author }: SendToKindleProps) {
  const [opened, setOpened] = useState(false);
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [email, setEmail] = useState('');
  const [emailError, setEmailError] = useState<string | null>(null);
  
  const dispatch = useAppDispatch();

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
    
    setLoading(true);
    setError(null);
    
    try {
      await dispatch(sendToKindle({
        book,
        email,
        title: author, // Swap: author field contains the actual title
        author: title  // Swap: title field contains the actual author
      })).unwrap();
      
      setSuccess(true);
      setTimeout(() => {
        setOpened(false);
        setSuccess(false);
        setLoading(false);
        setEmail('');
      }, 2000);
    } catch (err) {
      setError('Failed to send book to Kindle. Please try again.');
      setLoading(false);
    }
  };

  const handleClose = () => {
    setOpened(false);
    setLoading(false);
    setSuccess(false);
    setError(null);
    setEmailError(null);
    setEmail('');
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
        Send to Kindle
      </Button>

      <Modal
        opened={opened}
        onClose={handleClose}
        title={`Send "${author}" to Kindle`}
        centered
      >
        {success ? (
          <Stack align="center" spacing="md">
            <Text color="green" weight={500}>
              ✓ Book sent successfully!
            </Text>
            <Text size="sm" color="dimmed" align="center">
              The book has been sent to your Kindle email and will be deleted from the server.
            </Text>
          </Stack>
        ) : (
          <form onSubmit={handleSubmit}>
            <Stack spacing="md">
              <Text size="sm" color="dimmed">
                Enter your Kindle email address to receive "{author}" by {title}
              </Text>
              
              <TextInput
                label="Kindle Email Address"
                placeholder="your-kindle@kindle.com"
                required
                value={email}
                onChange={handleEmailChange}
                error={emailError}
                disabled={loading}
              />
              
              {error && (
                <Text color="red" size="sm">
                  {error}
                </Text>
              )}
              
              <Group position="right">
                <Button variant="subtle" onClick={handleClose} disabled={loading}>
                  Cancel
                </Button>
                <Button type="submit" loading={loading} color="green">
                  {loading ? 'Sending...' : 'Send to Kindle'}
                </Button>
              </Group>
            </Stack>
          </form>
        )}
      </Modal>
    </>
  );
}
