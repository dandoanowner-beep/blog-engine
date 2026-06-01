import { Link } from 'react-router-dom'

export default function GuestPrompt() {
  return (
    <div className="fixed inset-0 bg-gradient-to-t from-white via-white/90 to-transparent flex items-end justify-center pb-16 z-10" data-testid="guest-prompt">
      <div className="text-center px-4">
        <h3 className="text-xl font-semibold mb-2">Sign in to read the full post</h3>
        <p className="text-gray-500 mb-4">Join thousands of readers on BlogEngine</p>
        <div className="flex gap-3 justify-center">
          <Link
            to="/auth/register"
            className="bg-blue-600 text-white px-6 py-2 rounded-full font-medium hover:bg-blue-700"
          >
            Get started — free
          </Link>
          <Link
            to="/auth/login"
            className="border border-gray-300 px-6 py-2 rounded-full font-medium hover:bg-gray-50"
          >
            Sign in
          </Link>
        </div>
      </div>
    </div>
  )
}
