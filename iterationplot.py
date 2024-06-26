# reads csv file with new_max,iterations

import matplotlib.pyplot as plt
import pandas as pd

iteration_sizes = [3, 4, 5, 6, 7]
diff_mode = False

for size in iteration_sizes:
    # Read csv file
    df = pd.read_csv("data/size-{}-iterations.csv".format(size))
    if diff_mode:
        df['iterations'] = df['iterations'] - df['iterations'].shift(1)

    # Plot
    plt.plot(df['new_max'], df['iterations'], 'o-')
    plt.xlabel('new_max')
    plt.ylabel('iterations (diff)' if diff_mode else 'iterations')
    plt.title('Iterations{} vs new_max for size {}'.format(' diff' if diff_mode else '', size))

    plt.show()
