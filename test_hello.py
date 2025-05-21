import unittest
import io
import sys
from hello import main # Assuming hello.py has a main function

class TestHello(unittest.TestCase):

    def test_hello_world_output(self):
        captured_output = io.StringIO()
        sys.stdout = captured_output
        main()
        sys.stdout = sys.__stdout__  # Reset redirect.
        self.assertEqual(captured_output.getvalue().strip(), "Hello, World!")

if __name__ == '__main__':
    unittest.main()
