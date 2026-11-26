/* PageSpindle Metal availability probe. GPL-3.0-or-later.
 * No book input, paths, renderer arguments or persistent output. */
#import <Foundation/Foundation.h>
#import <Metal/Metal.h>
int main(void) { @autoreleasepool { return [MTLCopyAllDevices() count] ? 0 : 1; } }
