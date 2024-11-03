import { Button } from "~/components/Button.tsx";

export function App() {
  return (
    <>
      <div className="flex flex-col justify-center items-center gap-1 h-screen font-sans all:transition-100">
        <h2 className="text-5xl fw500">esm.sh</h2>
        <p className="text-gray-400 text-lg fw400">
          The <span className="fw600">no-build</span> cdn for modern web development.
        </p>
        <div className="mt2 flex justify-center text-2xl text-gray-400 hover:text-gray-700">
          <a
            className="i-carbon-logo-github text-inherit hover:animate-spin"
            href="https://github.com/esm-dev/esm.sh"
            target="_blank"
          />
        </div>
      </div>
      <div className="fixed bottom-5 w-full flex justify-center font-sans">
        <Button>Click Me</Button>
      </div>
    </>
  );
}
